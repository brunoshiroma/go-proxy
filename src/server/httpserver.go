package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func readConn(conn net.Conn) ([]byte, error) {
	fixedBuffer := make([]byte, 1024)

	read, err := conn.Read(fixedBuffer)
	if err != nil {
		log.Printf("ERROR readConn %v", err)
		return nil, err
	}

	bufferToWrite := make([]byte, read)
	copy(bufferToWrite, fixedBuffer)
	fixedBuffer = nil

	return bufferToWrite, nil
}

func pipe(source net.Conn, dest net.Conn) {

	source.SetDeadline(time.Now().Add(time.Second * 30))
	dest.SetDeadline(time.Now().Add(time.Second * 30))

	for {

		buffer := make([]byte, 1024)
		var read int
		var err error

		read, err = source.Read(buffer)

		if read <= 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		//log.Printf("Pipe read %v from %v", read, source.RemoteAddr())

		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error on reading from remote client %v", err)

				source.Close()
				dest.Close()

				return
			}
		} else {
			bufferToWrite := make([]byte, read)
			copy(bufferToWrite, buffer)
			_, err := dest.Write(bufferToWrite)

			bufferToWrite = nil
			buffer = nil

			//log.Printf("Pipe write %v to %v", read, dest.RemoteAddr())

			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("Error on writing on remote host %v", err)

					source.Close()
					dest.Close()

					return
				}
			}
		}

		buffer = nil

	}

}

func handleRedirect(req *http.Request, via []*http.Request) error {
	return errors.New("REDIRECT")
}

func handleHTTPRequest(conn net.Conn, requestString string) {
	stringParts := strings.SplitN(requestString, "\n", -1)
	stringConnect := strings.Split(stringParts[0], " ")

	defer conn.Close()

	request, err := http.NewRequest(stringConnect[0], stringConnect[1], nil)

	if err != nil {
		log.Printf("ERROR handleHttpRequest %v", err)
	} else {

		client := &http.Client{
			CheckRedirect: handleRedirect,
		}

		for _, parts := range stringParts[1 : len(stringParts)-1] {
			if strings.Index(parts, ":") > 0 {
				headerParts := strings.Split(parts, ": ")
				request.Header.Add(headerParts[0], strings.Trim(headerParts[1], "\r"))
			}
		}

		log.Printf("HTTP REQUEST %s", stringParts[0])
		response, err := client.Do(request)

		if err != nil {
			log.Printf("ERROR handleHttpRequest %v", err)

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
	bytes, err := readConn(conn)
	if err != nil {
		log.Printf("ERROR on handleConnection %v", err)
	} else {
		stringRequest := string(bytes)
		stringParts := strings.SplitN(stringRequest, "\n", -1)

		if strings.HasPrefix(stringParts[0], "CONNECT ") {
			stringConnect := strings.Split(stringParts[0], " ")
			log.Printf("CONNECT to %v", stringConnect[1])
			conn.Write([]byte("200 OK\r\n\r\n"))

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
				go handleConnection(conn)
			}

		}

	}
}
