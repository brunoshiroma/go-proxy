package server

import (
	"bufio"
	"net"
)

func newPipe(source net.Conn, dest net.Conn) {
	defer source.Close()
	defer dest.Close()
	sourceWriter := bufio.NewWriterSize(source, 512)
	sourceReader := bufio.NewReaderSize(source, 512)
	destWriter := bufio.NewWriterSize(dest, 512)
	destReader := bufio.NewReaderSize(dest, 512)

	sourceReader.WriteTo(destWriter)
	destReader.WriteTo(sourceWriter)
}
