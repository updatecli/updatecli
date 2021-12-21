#!/bin/sh
## This script updates the go version of a go mod file and prints the resulting content in the stdout
## Please note that it will NEVER change the go.mod file (works in a temp directory)
##
## Expected arguments:
## - 1: the path to the go.mod file
## - 2: the new golang version in semver format (major.minor.patch)

set -eux

go_mod_dir="$(dirname "${1}")"
go_version="${2}"
new_version="$(echo "${go_version}" | cut -d. -f1,2)"
tmp_dir="$(mktemp -d)"

## Ensures that there is a golang version installed
{
  if ! command -v go
  then
    curl --silent --show-error --location --output "${tmp_dir}/go.tgz" \
      "https://golang.org/dl/go${go_version}.$(uname -s | tr '[:upper:]' '[:lower:]')-amd64.tar.gz"
    mkdir -p "${tmp_dir}/.bin"
    tar xzf "${tmp_dir}/go.tgz" -C "${tmp_dir}/.bin"
    export PATH="${PATH}":"${tmp_dir}"/.bin/go/bin
  fi

  command -v go
  go version
} >&2

## Copy go mod's directory to a temp directory an start working from this temp. dir.
cp -r "${go_mod_dir}"/* "${tmp_dir}" >&2
cd "${tmp_dir}" >&2
GOPATH="$(mktemp -d)"
export GOPATH

## Update go mod properly
go mod edit -go="${new_version}" >&2
go mod tidy >&2
echo "" >> go.mod ## Ad empty endline to be POSIX compliant

## Show new go mod
cat go.mod
exit 0
