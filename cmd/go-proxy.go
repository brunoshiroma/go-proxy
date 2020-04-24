package main

import (
	"log"
	"os"
	"server"
	"strconv"
)

func main() {

	port, isSet := os.LookupEnv("PORT")

	if !isSet {
		port = "8080"
	}

	host, isSet := os.LookupEnv("HOST")

	if !isSet {
		host = "127.0.0.1"
	}

	uintPort, err := strconv.ParseUint(port, 10, 16)

	if err != nil {
		log.Fatalf("ERROR PORT %v", err)
	}

	log.Println("Starting go-proxy")
	server.InitHTTP(host, uint16(uintPort))
}
