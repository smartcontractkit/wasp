# Example Dockerfile for k8s run
# Builds all the tests in some directory that must have go.mod
# All tests are built as separate binaries with name "module.test"
FROM golang:1.21 as build
ARG TESTS_ROOT

WORKDIR /go/src
COPY . /tests

RUN CGO_ENABLED=0 cd /tests && go test -c ./...

FROM debian
ARG TESTS_ROOT

COPY --from=build /tests .
RUN apt-get update && apt-get install -y ca-certificates
ENTRYPOINT /bin/bash