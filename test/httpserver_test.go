package server

import (
	"testing"

	"github.com/brunoshiroma/go-proxy/internal/server"
)

func TestReadConn(t *testing.T) {
	bytes, error := server.ReadConn(nil)
	if error != nil {
		t.Errorf(error.Error())
	} else if bytes != nil {
		t.Errorf("shold be nil, but is %v", bytes)
	}
}
