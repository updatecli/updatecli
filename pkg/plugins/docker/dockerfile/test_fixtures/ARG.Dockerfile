FROM golang:1.15 AS builder

ARG HELM_VERSION=3.0.0

WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...

ENV HELM_VERSION "${HELM_VERSION}"

RUN curl -ssLO https://helm.sh/downloads/${HELM_VERSION}.tgz

FROM golang:1.15 as tester

arg HELM_VERSION=3.0.0

ENV HELM_VERSION="${HELM_VERSION}"

RUN curl -ssLO https://helm.sh/downloads/${HELM_VERSION}.tgz

RUN go test ./...

FROM golang AS reporter

ARG HELM_VERSION

RUN curl -ssLO https://helm.sh/downloads/${HELM_VERSION}.tgz

FROM golang

RUN echo OK

FROM ubuntu AS base

ARG helm_version

FROM ubuntu AS golang

FROM ubuntu:20.04

arg helm_version

LABEL maintainer="Olblak <me@olblak.com>"

VOLUME /tmp

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

RUN \
  apt-get update && \
  apt-get install -y ca-certificates && \
  apt-get clean && \
  find /var/lib/apt/lists -type f -delete

USER updatecli

WORKDIR /home/updatecli

COPY --from=builder --chown=updatecli:updatecli /go/src/app/dist/updatecli /usr/bin/updatecli

ENTRYPOINT [ "/usr/bin/updatecli" ]

CMD ["--help"]
