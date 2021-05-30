package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type redirect interface {
	Redirect() (int, string)
}

type redirectError struct {
	status   int
	location string
}

func (e *redirectError) Error() string {
	return fmt.Sprintf("%v %v", e.status, e.location)
}

/*
ReadConn read all the data on conn, and return the bytes
*/
func ReadConn(conn net.Conn) ([]byte, error) {

	if conn == nil {
		return nil, nil
	}

	fixedBuffer := make([]byte, 1024*10)

	read, err := conn.Read(fixedBuffer)
	if err != nil {
		log.Printf("ERROR readConn %v", err)
		fixedBuffer = nil
		return nil, err
	}

	bufferToWrite := make([]byte, read)
	copy(bufferToWrite, fixedBuffer)
	fixedBuffer = nil

	return bufferToWrite, nil
}

func pipe(source net.Conn, dest net.Conn) {
	var (
		read int
		err  error
	)
	source.SetDeadline(time.Now().Add(time.Second * 60))
	dest.SetDeadline(time.Now().Add(time.Second * 60))

	defer source.Close()
	defer dest.Close()

	for {

		buffer := make([]byte, 1024*10)

		read, err = source.Read(buffer)

		if err != nil {
			if err == io.EOF {
				log.Printf("INFO EOF")
			} else if strings.Contains(err.Error(), "poll.DeadlineExceededError") {
				log.Printf("INFO timeout")
			} else {
				log.Printf("ERROR on reading from tcp %#v src=%#v dst=%#v", err, source.RemoteAddr().String(), dest.RemoteAddr().String())
			}
			break
		}

		bufferToWrite := make([]byte, read)
		copy(bufferToWrite, buffer)
		_, err := dest.Write(bufferToWrite)

		bufferToWrite = nil
		buffer = nil

		if err != nil {
			if err != io.EOF {
				log.Printf("Error on writing on remote host %v", err)
				return
			}
			source.Close()
			dest.Close()
		}

	}

}

func handleRedirect(req *http.Request, via []*http.Request) error {
	//handling redirect
	if req.Response.StatusCode == 301 || req.Response.StatusCode == 302 {
		location := req.Response.Header["Location"]
		if location != nil && len(location) > 0 {
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

func handleHTTPRequest(conn net.Conn, requestString string) {
	stringParts := strings.SplitN(requestString, "\n", -1)
	stringConnect := strings.Split(stringParts[0], " ")

	stringRequestContentParts := strings.SplitN(requestString, "\r\n\r\n", -1) // request content most have 2 new lines

	defer conn.Close()

	var request *http.Request = nil
	var err error = nil

	if len(stringRequestContentParts) > 1 {
		stringRequestContent := stringRequestContentParts[1]
		request, err = http.NewRequest(stringConnect[0], stringConnect[1], strings.NewReader(stringRequestContent))
	} else {
		request, err = http.NewRequest(stringConnect[0], stringConnect[1], nil)
	}

	if err != nil {
		log.Printf("ERROR handleHttpRequest %v", err)
	} else {

		client := &http.Client{
			CheckRedirect: handleRedirect,
		}

		for _, parts := range stringParts[1 : len(stringParts)-1] {
			if strings.Index(parts, ":") > 0 {
				headerParts := strings.Split(parts, ": ")
				if len(headerParts) > 1 {
					request.Header.Add(headerParts[0], strings.Trim(headerParts[1], "\r"))
				}
			}
		}

		log.Printf("HTTP REQUEST %s", stringParts[0])
		response, err := client.Do(request)

		if err != nil {

			if data, ok := err.(*url.Error); ok {
				if e, ok := data.Err.(*redirectError); ok {
					conn.Write([]byte(fmt.Sprintf("%s %d\r\n", "HTTP/1.0", e.status)))
					conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", "Location", e.location)))
				} else {
					log.Printf("ERROR handleHttpRequest %v", err)
				}
			} else {
				log.Printf("ERROR handleHttpRequest %v", err)
			}

		} else {

			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)

			if err != nil {
				log.Printf("ERROR handleHttpRequest %v", err)
			} else {

				conn.Write([]byte(fmt.Sprintf("%s %d\r\n", response.Proto, response.StatusCode)))

				for header := range response.Header {
					conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", header, response.Header.Get(header))))
				}

				conn.Write([]byte("\r\n"))
				conn.Write(body)

			}

		}

	}

}

func handleConnection(conn net.Conn) {
	bytes, err := ReadConn(conn)
	if err != nil {
		log.Printf("ERROR on handleConnection %v", err)
		conn.Close()
	} else {
		stringRequest := string(bytes)
		stringParts := strings.SplitN(stringRequest, "\n", -1)

		if strings.HasPrefix(stringParts[0], "CONNECT ") {
			stringConnect := strings.Split(stringParts[0], " ")
			log.Printf("CONNECT to %v", stringConnect[1])
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

			bytes = nil
			stringParts = nil

			handleProxyConn(conn, stringConnect[1])
		} else {
			handleHTTPRequest(conn, stringRequest)
		}
	}

}

func handleProxyConn(source net.Conn, dest string) {
	remoteConn, err := net.DialTimeout("tcp", dest, time.Second*30)

	if err != nil {
		log.Printf("ERROR handleProxyConn %v", err)
	} else {
		go pipe(source, remoteConn)
		go pipe(remoteConn, source)
	}
}

/*InitHTTP start the http proxy server
 */
func InitHTTP(host string, port uint16) {

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
				defer conn.Close()

				go handleConnection(conn)
			}

		}

	}
}
