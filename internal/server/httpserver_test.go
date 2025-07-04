package server

import (
	"net"
	"testing"
)

func TestHttpServer(t *testing.T) {
	httpServer := HttpServer{}

	t.Run("Test ReadConn nil conn success", func(t *testing.T) {
		bytes, error := httpServer.ReadConn(nil)
		if error != nil {
			t.Errorf("Error %s", error.Error())
		} else if bytes != nil {
			t.Errorf("shold be nil, but is %v", bytes)
		}
	})

	t.Run("Test httpRequest success", func(t *testing.T) {
		client, server := net.Pipe()
		httpServer.handleHTTPRequest(client, "GET /\n")
		httpServer.handleHTTPRequest(server, "POST /\n")
	})

}
