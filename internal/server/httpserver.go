package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

/*
HttpServer server proxy the connections.
Its handles http and https proxy connections
*/
type HttpServer struct {
	useNewPipe     bool
	pipe           func(net.Conn, net.Conn)
	serverHostName string
}

type redirectError struct {
	status   int
	location string
}

const httpHandleErrorStr = "ERROR handleHttpRequest %v"

var httpClient = &http.Client{
	CheckRedirect: handleRedirect,
}

func (e *redirectError) Error() string {
	return fmt.Sprintf("%v %v", e.status, e.location)
}

/*
ReadConn read all the data on conn, and return the bytes
*/
func (s *HttpServer) ReadConn(conn net.Conn) ([]byte, error) {

	if conn == nil {
		return nil, nil
	}

	fixedBuffer := make([]byte, 1024*10)

	read, err := conn.Read(fixedBuffer)
	if err != nil {
		log.Printf("ERROR readConn %v", err)
		return nil, err
	}

	bufferToWrite := make([]byte, read)
	copy(bufferToWrite, fixedBuffer)

	return bufferToWrite, nil
}

func handleRedirect(req *http.Request, via []*http.Request) error {
	//handling redirect
	if req.Response.StatusCode == 301 || req.Response.StatusCode == 302 {
		location := req.Response.Header["Location"]
		if len(location) > 0 {
			newLocation := location[0]
			return &redirectError{
				status:   req.Response.StatusCode,
				location: newLocation,
			}
		}
	}

	if len(via) > 2 {
		return fmt.Errorf("max 2 hops")
	}
	return nil
}

func (s *HttpServer) handleHTTPRequest(conn net.Conn, requestString string) {

	defer conn.Close()

	request, err := s.parseHttpRequestString(requestString)

	if err != nil {
		log.Printf(httpHandleErrorStr, err)
		return
	}

	if isHealthCheck(request, s.serverHostName) {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return
	}

	log.Printf("HTTP REQUEST %s", request.URL)
	response, err := httpClient.Do(request)

	if err != nil {
		handleRedirectError(err, conn)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Printf(httpHandleErrorStr, err)
		return
	}

	conn.Write([]byte(fmt.Sprintf("%s %d\r\n", response.Proto, response.StatusCode)))

	for header := range response.Header {
		conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", header, response.Header.Get(header))))
	}

	conn.Write([]byte("\r\n"))
	conn.Write(body)
}

func isHealthCheck(request *http.Request, serverHostName string) bool {
	host := request.Header.Get("Host")
	log.Printf("Checking if request is healthcheck host header %s, method %s, Path %s", host, request.Method, request.URL.Path)
	return serverHostName == host && request.Method == "GET" && request.URL.Path == "/health" ||
		serverHostName == host && request.Method == "HEAD" && request.URL.Path == "/"
}

func handleRedirectError(err error, conn net.Conn) {
	if data, ok := err.(*url.Error); ok {
		if e, ok := data.Err.(*redirectError); ok {
			conn.Write([]byte(fmt.Sprintf("%s %d\r\n", "HTTP/1.0", e.status)))
			conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", "Location", e.location)))
		} else {
			log.Printf(httpHandleErrorStr, err)
		}
	} else {
		log.Printf(httpHandleErrorStr, err)
	}
}

func (s *HttpServer) parseHttpRequestString(requestString string) (request *http.Request, err error) {
	stringParts := strings.SplitN(requestString, "\n", -1)
	stringConnect := strings.Split(stringParts[0], " ")

	if len(stringConnect) <= 1 {
		err = fmt.Errorf("INVALID REQUEST STRING %v", requestString)
		return
	}

	stringRequestContentParts := strings.SplitN(requestString, "\r\n\r\n", -1) // request content most have 2 new lines
	if len(stringRequestContentParts) > 1 {
		stringRequestContent := stringRequestContentParts[1]
		request, err = http.NewRequest(stringConnect[0], stringConnect[1], strings.NewReader(stringRequestContent))
	} else {
		request, err = http.NewRequest(stringConnect[0], stringConnect[1], nil)
	}

	if err != nil {
		return
	}

	for _, parts := range stringParts[1 : len(stringParts)-1] {
		if strings.Index(parts, ":") > 0 {
			headerParts := strings.Split(parts, ": ")
			if len(headerParts) > 1 {
				request.Header.Add(headerParts[0], strings.Trim(headerParts[1], "\r"))
			}
		}
	}
	return
}

func (s *HttpServer) handleConnection(conn net.Conn) {
	defer handlePanic()
	bytes, err := s.ReadConn(conn)
	if err != nil {
		log.Printf("ERROR on handleConnection %v", err)
		conn.Close()
	} else {
		stringRequest := string(bytes)
		stringParts := strings.SplitN(stringRequest, "\n", -1)

		if strings.HasPrefix(stringParts[0], "CONNECT ") { //https
			stringConnect := strings.Split(stringParts[0], " ")
			log.Printf("CONNECT to %v", stringConnect[1])
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

			s.handleProxyConn(conn, stringConnect[1])
		} else { // http
			s.handleHTTPRequest(conn, stringRequest)
		}
	}

}

func (s *HttpServer) handleProxyConn(source net.Conn, dest string) {
	remoteConn, err := net.DialTimeout("tcp", dest, time.Second*30)

	if err != nil {
		log.Printf("ERROR handleProxyConn %v", err)
	} else {
		go s.pipe(source, remoteConn)
		go s.pipe(remoteConn, source)
	}
}

/*InitHTTP start the http proxy server
 */
func (s *HttpServer) InitHTTP(host string, port uint16, useNewPipe bool, serverHostName string) {
	s.useNewPipe = useNewPipe
	s.serverHostName = serverHostName

	if useNewPipe {
		s.pipe = newPipe
		log.Println("using the new pipe")
	} else {
		s.pipe = pipe
		log.Println("using the old pipe")
	}

	l, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Printf("ERROR ON LISTEN %v", err)
	} else {

		log.Printf("Listening on %s:%d", host, port)

		defer l.Close()

		for {
			conn, err := l.Accept()

			if err != nil {
				log.Printf("ERROR ON ACCEPT %v", err)
			} else {
				go s.handleConnection(conn)
			}

		}

	}
}

func handlePanic() {
	if panicError := recover(); panicError != nil {
		log.Printf("PANIC RECOVER %s\n", panicError)
		log.Println(string(debug.Stack()))
	}
}
