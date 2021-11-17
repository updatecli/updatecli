FROM golang:1.15 AS builder

ARG golang=3.0.0

WORKDIR /go/src/app

COPY ./golang .

RUN go get -d -v ./... && echo golang

FROM golang:1.15 as tester

RUN go test ./...

FROM golang AS reporter

RUN go tool cover ./..

FROM golang

RUN echo "${GOLANG}"

FROM ubuntu AS base

RUN apt-get update

FROM ubuntu AS golang

RUN apt-get update

FROM ubuntu:20.04

RUN apt-get update

LABEL maintainer="golang"
LABEL golang="${GOLANG_VERSION}"

VOLUME /tmp / golang

USER golang

WORKDIR /home/updatecli

COPY --from=golang --chown=updatecli:golang /go/src/app/dist/updatecli /usr/bin/golang

ENTRYPOINT [ "/usr/bin/golang" ]

CMD ["--help:golang"]
