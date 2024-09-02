FROM golang:1.15 AS builder

LABEL org.opencontainers.image.version=0.14.0

FROM golang:1.15 as tester

label org.opencontainers.image.version=0.14.0

FROM golang AS reporter

LABEL org.opencontainers.image.version

FROM golang

RUN echo OK

FROM ubuntu AS base

LABEL org.opencontainers.image.version
label org.opencontainers.image.version=0.14.0

FROM ubuntu AS golang
