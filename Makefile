build:
	go build -o go-proxy cmd/go-proxy/main.go

run-k6-docker-http:
	docker run --rm -v "`pwd`/k6/http.js:/http.js" -e HTTP_PROXY=127.0.0.1:8080 --net=host loadimpact/k6 run /http.js

run-k6-docker-https:
	docker run --rm -v "`pwd`/k6/https.js:/https.js" -e HTTPS_PROXY=http://127.0.0.1:8080 --net=host loadimpact/k6 run /https.js

run-k6-docker-all: run-k6-docker-http run-k6-docker-https

clean:
	go clean -cache -testcache
	rm go-proxy
	rm cmd/go-proxy/__debug_bin
	rm main

deep-clean: clean
	go clean -modcache

test:
	go test -cover -coverprofile=coverage.out ./...

test-with-report: test
	go tool cover -html=coverage.out
