ARG GORELEASER_VERSION=0.156.2
FROM goreleaser/goreleaser:v${GORELEASER_VERSION} as goreleaser
FROM golang:1.16 as builder

COPY --from=goreleaser /usr/local/bin/goreleaser /usr/local/bin/goreleaser

WORKDIR /go/src/app

COPY . /go/src/app

## Default make build is a "dirty/snapshot" build
ARG MAKE_TARGET=build.all
RUN make "${MAKE_TARGET}"

###

FROM ubuntu:20.04

LABEL maintainer="Olblak <me@olblak.com>"

VOLUME /tmp

COPY --from=builder --chown=updatecli:updatecli /go/src/app/dist/updatecli_*.deb /tmp/updatecli_*.deb

RUN apt-get update && \
    apt-get install --no-install-recommends  -y ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    apk add --no-cache --allow-untrusted /tmp/updatecli_*.apk

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

USER updatecli

WORKDIR /home/updatecli

ENTRYPOINT [ "/usr/bin/updatecli" ]

CMD ["--help"]
