ARG alpine_version=3.16.3
ARG debian_version 8
ARG platform linux/arch64
FROM --platform=linux/ppc64 alpine:${alpine_version} AS alpine
FROM --platform=${platform} debian:${debian_version}
FROM --platform=windows/ppc64 opensuse:15.4
# Using multiple variable from a Dockerfile instruction is currently not supported by Updatecli, and therefor ignore it
FROM --platform=linux/ppc64 alpine:${alpine_version}${debian_version} AS alpine