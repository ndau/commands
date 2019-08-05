#!/bin/bash

NDEV_SUBDIR=github.com/oneiro-ndev
NDEV_DIR="$GOPATH/src/$NDEV_SUBDIR"

BIN_DIR=/image/bin
mkdir "$BIN_DIR"

echo Building ndau...
cd "$NDEV_DIR"/commands || exit 1
VERSION=$(git describe --long --tags --match="v*")
echo "  VERSION=$VERSION"
VERSION_PKG="$NDEV_SUBDIR/commands/vendor/$NDEV_SUBDIR/ndau/pkg/version"
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndaunode
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndauapi
mv ndaunode "$BIN_DIR"
mv ndauapi "$BIN_DIR"

echo Building generate...
go build ./cmd/generate
mv generate "$BIN_DIR"

echo Building procmon...
go build ./cmd/procmon
mv procmon "$BIN_DIR"

echo Setup complete
