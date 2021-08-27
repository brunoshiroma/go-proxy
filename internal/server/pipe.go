package server

import (
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const tcp_timeout_secs int32 = 120

func pipe(source net.Conn, dest net.Conn) {
	var (
		read int
		err  error
	)
	source.SetDeadline(time.Now().Add(time.Second * time.Duration(tcp_timeout_secs)))

	defer source.Close()
	defer dest.Close()
	buffer := make([]byte, 64)

	for {

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

		if read <= 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		source.SetDeadline(time.Now().Add(time.Second * time.Duration(tcp_timeout_secs)))

		bufferToWrite := make([]byte, read)
		copy(bufferToWrite, buffer)
		dest.SetDeadline(time.Now().Add(time.Second * time.Duration(tcp_timeout_secs)))
		_, err := dest.Write(bufferToWrite)

		if err != nil {
			if err != io.EOF {
				log.Printf("Error on writing on remote host %v", err)
				return
			}
			source.Close()
			dest.Close()
			break
		}

	}

}
