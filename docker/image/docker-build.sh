#!/bin/bash

set -e

NDEV_SUBDIR=github.com/ndau
NDEV_DIR="$GOPATH/src/$NDEV_SUBDIR"

BIN_DIR=/image/bin
mkdir "$BIN_DIR"

cd "$NDEV_DIR"/commands || exit 1

echo Building ndau...
VERSION=$(git describe --long --tags --match="v*")
echo "  VERSION=$VERSION"
VERSION_PKG="$NDEV_SUBDIR/commands/vendor/$NDEV_SUBDIR/ndau/pkg/version"
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndaunode
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndauapi
mv ndaunode "$BIN_DIR"
mv ndauapi "$BIN_DIR"

echo Building tools...
go build ./cmd/generate
go build ./cmd/keytool
go build ./cmd/ndau
mv generate "$BIN_DIR"
mv keytool "$BIN_DIR"
mv ndau "$BIN_DIR"

echo Building procmon...
go build ./cmd/procmon
mv procmon "$BIN_DIR"

if [ -n "$RUN_UNIT_TESTS" ]; then
    echo "Running unit tests..."
    export CGO_ENABLED=0
    for dir in "$NDEV_DIR"/commands/vendor/"$NDEV_SUBDIR"/*
    do
        cd "$dir"
        pwd
        go test ./...
    done
fi
