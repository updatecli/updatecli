## Retrieve goreleaser binary
ARG GORELEASER_VERSION=0.183.0
ARG BUILDPLATFORM=amd64
ARG GOLANG_VERSION=1.17.2

## Goreleaser requires the "--platform" to handle multi-platform builds
# hadolint ignore=DL3029
FROM --platform=${BUILDPLATFORM} goreleaser/goreleaser:v${GORELEASER_VERSION} as goreleaser

## Build using golang docker image
## Goreleaser requires the "--platform" to handle multi-platform builds
# hadolint ignore=DL3029
FROM --platform=${BUILDPLATFORM} golang:${GOLANG_VERSION} as builder

WORKDIR /go/src/app

COPY --from=goreleaser /usr/local/bin/goreleaser /usr/local/bin/goreleaser
COPY . .

ARG MAKE_TARGET=build
RUN make "${MAKE_TARGET}"

## Build final updatecli docker image
## Goreleaser requires the "--platform" to handle multi-platform builds
# hadolint ignore=DL3029
FROM --platform=${BUILDPLATFORM} debian:stable-slim

LABEL maintainer="Olblak <me@olblak.com>"

ARG TARGETOS=linux
ARG TARGETARCH=amd64

# /tmp is used by updatecli to store git repository so it's better
# to define a volume
VOLUME /tmp

## The latest version of these "generic" package is always required
# hadolint ignore=DL3008
RUN apt-get update && \
    apt-get install --yes --no-install-recommends \
        ca-certificates \
        curl \
        tar \
        unzip \
    wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/src/app/dist/updatecli_${TARGETOS}_${TARGETARCH}/updatecli /usr/local/bin/updatecli

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

USER updatecli

WORKDIR /home/updatecli

ENTRYPOINT [ "/usr/local/bin/updatecli" ]
CMD ["help"]
