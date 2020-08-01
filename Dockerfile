FROM --platform=$TARGETPLATFORM golang:alpine AS build-base
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN apk add build-base

FROM build-base as build
ARG TARGETPLATFORM
ARG BUILDPLATFORM
WORKDIR /proxy
COPY . .
#because of go.mod
RUN unset GOPATH
RUN CGO_ENABLED=0 GOOS=linux go build -o go-proxy -a -ldflags '-extldflags "-static"' cmd/go-proxy/main.go


FROM --platform=$TARGETPLATFORM alpine AS runtime
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

WORKDIR /proxy
COPY --from=build /proxy/go-proxy .
ENTRYPOINT ["./go-proxy"]