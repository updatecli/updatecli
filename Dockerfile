FROM golang:1.15 as builder

# RUN go get -d -v ./...

ARG GORELEASER_VERSION=0.156.2
RUN curl --silent --show-error --location --output "/tmp/goreleaser.tgz" \
  "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz" \
  && tar xzf /tmp/goreleaser.tgz --directory /usr/local/bin goreleaser \
  && goreleaser --version 2>&1 | grep -q "${GORELEASER_VERSION}" \
  && rm -f /tmp/goreleaser.tgz

WORKDIR /go/src/app

COPY . .
RUN make build.all

###

FROM ubuntu:20.04

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
COPY --from=builder --chown=updatecli:updatecli /go/src/app/dist/updatecli_linux_amd64/updatecli /usr/bin/updatecli

ENTRYPOINT [ "/usr/bin/updatecli" ]

CMD ["--help"]
