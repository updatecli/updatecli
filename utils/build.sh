#!/bin/bash

OS_TARGETS="${OS_TARGETS:-"darwin linux windows"}"
ARCH_TARGETS="${ARCH_TARGETS:-"amd64 386"}"

BUILD_DATE=$(date -R)
VERSION=$(git describe --tags)
GOVERSION=$(go version)

BASENAME="updatecli"

echo "Building updatecli OS packages for version $VERSION"

for GOOS in $OS_TARGETS; do
  for GOARCH in $ARCH_TARGETS; do

    if [ -z "${CUSTOM_BINARY}" ]
    then
      BINARY="$BASENAME.$GOOS.$GOARCH"
      if [ $GOOS = "windows" ]; then
          BINARY="$BASENAME.$GOARCH"
          BINARY+='.exe'
      fi
      BINARY="./bin/${BINARY}"
    else
      BINARY="${CUSTOM_BINARY}"
    fi

    echo "Build $BINARY_DIRECTORY/$BINARY for $GOOS-$GOARCH"

    GOOS=$GOOS GOARCH=$GOARCH go build \
      -ldflags "-w -s \
        -X \"github.com/olblak/updateCli/pkg/core/version.BuildTime=$BUILD_DATE\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.GoVersion=$GOVERSION\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.Version=$VERSION\""\
      -o "$BINARY"
  done;
done
