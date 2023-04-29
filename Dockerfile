# Example Dockerfile for k8s run
FROM golang:1.20 as build
ARG BUILD_ROOT
ARG TEST_NAME

WORKDIR /go/src
COPY . .

RUN CGO_ENABLED=0 cd ${BUILD_ROOT} && go test -c -o wasp_test

FROM debian
ARG BUILD_ROOT
ARG TEST_NAME
ENV test_name=$TEST_NAME

COPY --from=build ${BUILD_ROOT} /
ENTRYPOINT ./wasp_test -test.run $test_name -test.timeout 24h