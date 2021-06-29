package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"

	"github.com/brunoshiroma/go-proxy/internal/server"
)

func main() {
	var useNewPipe bool
	log.Printf("Starting go-proxy on %s/%s", runtime.GOOS, runtime.GOARCH)

	port, isSet := os.LookupEnv("PORT")

	if !isSet {
		port = "8080"
	}

	host, isSet := os.LookupEnv("HOST")

	if !isSet {
		host = "127.0.0.1"
	}

	debug, isSet := os.LookupEnv("GO_PROXY_PPROF_DEBUG")

	if isSet && debug == "true" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	uintPort, err := strconv.ParseUint(port, 10, 16)

	useNewPipeStr, isSet := os.LookupEnv("NEW_PIPE")

	if isSet && useNewPipeStr == "true" {
		useNewPipe = true
	}

	if err != nil {
		log.Fatalf("ERROR PORT %v", err)
	}

	htpServer := server.HttpServer{}
	htpServer.InitHTTP(host, uint16(uintPort), useNewPipe)
}
