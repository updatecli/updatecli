### File: Dockerfile.webapp
ARG GO_VERSION=1.15.8
FROM golang:"${GO_VERSION}-alpine" AS build
# Retrieve source code
WORKDIR /app
COPY . /app
# Build application
RUN go build -X "GoVersion=${GO_VERSION}" -o ./webapp

FROM ubuntu AS test
ARG GO_VERSION
ENV GO_VERSION=${GO_VERSION}
RUN ./run_tests.sh

FROM ubuntu:18.04 AS run
COPY --from=builder /app/webapp /usr/local/bin/webapp

FROM jenkins/jenkins:2.274
ENV JAVA_VERSION=8
LABEL maintainer=olblak
LABEL version=2.274  \
  releasedate="2021/01/09"
