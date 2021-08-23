## Retrieve goreleaser binary
ARG GORELEASER_VERSION=0.156.2
ARG BUILDPLATFORM=amd64

FROM --platform=${BUILDPLATFORM} goreleaser/goreleaser:v${GORELEASER_VERSION} as goreleaser

## Build using golang docker image
FROM --platform=${BUILDPLATFORM} golang:1.17.0 as builder

WORKDIR /go/src/app

COPY --from=goreleaser /usr/local/bin/goreleaser /usr/local/bin/goreleaser
COPY . .

ARG MAKE_TARGET=build
RUN make "${MAKE_TARGET}"

## Build final updatecli docker image
FROM --platform=${BUILDPLATFORM} debian:stable-slim

LABEL maintainer="Olblak <me@olblak.com>"

ARG TARGETOS=linux
ARG TARGETARCH=amd64

# /tmp is used by updatecli to store git repository so it's better
# to define a volume
VOLUME /tmp

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    apt-get clean && \
    find /var/lib/apt/lists -type f -delete

COPY --from=builder /go/src/app/dist/updatecli_${TARGETOS}_${TARGETARCH}/updatecli /usr/local/bin/updatecli

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

USER updatecli

WORKDIR /home/updatecli

ENTRYPOINT [ "/usr/local/bin/updatecli" ]
CMD ["help"]
