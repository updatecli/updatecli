FROM golang:1.24.3 AS normal-upper-case-equal

ENV TERRAFORM_VERSION=0.14.0

FROM golang:1.24.3 AS normal-upper-case-space

ENV TERRAFORM_VERSION "0.14.0"

FROM golang:1.24.3 AS normal-lower-case-equal

env TERRAFORM_VERSION=0.14.0

FROM golang:1.24.3 AS no-catch-lower-case

ENV terraform_version=0.14.0

FROM golang:1.24.3 AS no-catch-same-prefix-equal

ENV TERRAFORM=bar

FROM golang:1.24.3 AS no-catch-same-prefix-space

ENV TERRAFORM "bar"

FROM golang:1.24.3 AS default-multi-instructions-first

ENV TERRAFORM_VERSION=0.14.0 \
  TERRAFORM=bar \
  enable_feature="true"

FROM golang:1.24.3 AS default-multi-instructions-other

ENV TERRAFORM=bar \
  TERRAFORM_VERSION=0.14.0 \
  enable_feature="true"
