# Example Dockerfile for k8s run
FROM golang:1.20 as build
ARG BUILD_ROOT

WORKDIR /go/src
COPY . .

RUN CGO_ENABLED=0 cd ${BUILD_ROOT} && go test -c -o wasp_test

FROM debian
ARG BUILD_ROOT

COPY --from=build ${BUILD_ROOT} /
ENTRYPOINT /bin/bash