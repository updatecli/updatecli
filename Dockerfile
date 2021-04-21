## Retrieve goreleaser binary
ARG GORELEASER_VERSION=0.156.2
ARG BUILDPLATFORM=amd64

FROM --platform=${BUILDPLATFORM} goreleaser/goreleaser:v${GORELEASER_VERSION} as goreleaser

## Build using golang docker image
FROM --platform=${BUILDPLATFORM} golang:1.16 as builder

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

COPY --from=builder /go/src/app/dist/updatecli_${TARGETOS}_${TARGETARCH}/updatecli /usr/local/bin/updatecli

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

USER updatecli

WORKDIR /home/updatecli

ENTRYPOINT [ "/usr/local/bin/updatecli" ]
CMD ["help"]
