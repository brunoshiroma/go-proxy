[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=brunoshiroma_go-proxy&metric=alert_status)](https://sonarcloud.io/dashboard?id=brunoshiroma_go-proxy)
[![Build Status](https://travis-ci.com/brunoshiroma/go-proxy.svg?branch=master)](https://travis-ci.com/brunoshiroma/go-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/brunoshiroma/go-proxy)](https://goreportcard.com/report/github.com/brunoshiroma/go-proxy)

# Simple HTTPS Proxy - Written in Go
Developed with go
```
go version go1.23.0 linux/amd64
```

Simple HTTPS proxy, using CONNECT pattern

And with Goroutines =)

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbrunoshiroma%2Fgo-proxy.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbrunoshiroma%2Fgo-proxy?ref=badge_large)

## USAGE

Simple run the proxy
```
with binary
./go-proxy
OR from source
go run cmd/go-proxy/main.go
```

## Stress test
Stress tests using [k6](https://k6.io), tests are located on folder k6.
Running with docker ( with make )
```bash
make run-k6-docker-http
# OR
make run-k6-docker-https
```

### ENV VARS
```bash
GO_PROXY_PPROF_DEBUG=true #enable the PPROF profiling on 127.0.0.1:6060
HOST=0.0.0.0 #set the host/ip to bind the listening proxy address
PORT=8080 #set the port for the binding
```
# pre-commit
 * [pre-commit](https://pre-commit.com/)

# Code Quality
 * [golangci-lint](https://github.com/golangci/golangci-lint)
