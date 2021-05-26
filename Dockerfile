FROM golang:alpine AS build-base

RUN apk add build-base

FROM build-base as build
WORKDIR /proxy
COPY . .
#because of go.mod
RUN unset GOPATH
RUN CGO_ENABLED=0 GOOS=linux go build -o go-proxy -a -ldflags '-extldflags "-static"' cmd/go-proxy/main.go


FROM alpine AS runtime

WORKDIR /proxy
COPY --from=build /proxy/go-proxy .
ENTRYPOINT ["./go-proxy"]