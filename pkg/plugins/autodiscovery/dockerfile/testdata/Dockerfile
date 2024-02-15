FROM --platform=${BUILDPLATFORM} updatecli/updatecli:v0.37.0
ENV maintainer=updatecli.io
#---
FROM updatecli/updatecli:v0.38.0

FROM updatecli/updatecli:v0.36.0 as builder

ARG alpine_version=3.16.3
FROM alpine:${alpine_version} AS base
