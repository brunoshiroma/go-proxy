package server

import (
	"net"
	"sync"
	"testing"
)

func TestNewPipe(t *testing.T) {
	t.Run("New pipe success", func(t *testing.T) {
		client, remote := net.Pipe()
		wg := sync.WaitGroup{}
		//Wbuffer := make([]byte, 3)
		wg.Add(1)
		go func() {
			newPipe(client, remote)
			wg.Done()
		}()
		t.Log("Writing")
		//client.Write([]byte{1, 2, 3})
		//remote.Read(buffer)
		client.Close()
		remote.Close()
		wg.Wait()
	})
}
