[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=brunoshiroma_go-proxy&metric=alert_status)](https://sonarcloud.io/dashboard?id=brunoshiroma_go-proxy)
[![Build Status](https://travis-ci.com/brunoshiroma/go-proxy.svg?branch=master)](https://travis-ci.com/brunoshiroma/go-proxy)


# Simple HTTPS Proxy - Written in Go
Developed with go
```
go version go1.15.3 linux/amd64
```

Simple HTTPS proxy, using CONNECT pattern

And with Goroutines =)

## USAGE

Simple run the proxy
```
with binary
./go-proxy
OR from source
go run cmd/go-proxy/main.go
```

### ENV VARS
```bash
GO_PROXY_PPROF_DEBUG=true #enable the PPROF profiling on 127.0.0.1:6060
HOST=0.0.0.0 #set the host/ip to bind the listening proxy address
PORT=8080 #set the port for the binding
```