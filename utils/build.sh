#!/bin/bash

BUILD_DATE=$(date -R)
VERSION=$(git describe --tags)
GOVERSION=$(go version)

BASENAME="updatecli"

echo "Building updatecli OS packages for version $VERSION"

for GOOS in darwin linux windows; do
  for GOARCH in amd64 386; do

    BINARY="$BASENAME.$GOOS.$GOARCH"
    BINARY_DIRECTORY="./bin"

    if [ $GOOS = "windows" ]; then
        BINARY="$BASENAME.$GOARCH"
        BINARY+='.exe'

    fi

    echo "Build $BINARY_DIRECTORY/$BINARY for $GOOS-$GOARCH"

    GOOS=$GOOS GOARCH=$GOARCH go build \
      -ldflags "-w -s \
        -X \"github.com/olblak/updateCli/pkg/core/version.BuildTime=$BUILD_DATE\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.GoVersion=$GOVERSION\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.Version=$VERSION\""\
      -o "$BINARY_DIRECTORY/$BINARY"
  done;
done
