FROM --platform=$TARGETPLATFORM golang:alpine AS build-base
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN apk add build-base

FROM build-base as build
ARG TARGETPLATFORM
ARG BUILDPLATFORM
WORKDIR /proxy
COPY . .
RUN GOPATH=/proxy CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' cmd/go-proxy/


FROM --platform=$TARGETPLATFORM alpine AS runtime
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

WORKDIR /proxy
COPY --from=build /proxy/go-proxy .
ENTRYPOINT ["./go-proxy"]