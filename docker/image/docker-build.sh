#!/bin/bash

set -e

NDEV_SUBDIR=github.com/oneiro-ndev
NDEV_DIR="$GOPATH/src/$NDEV_SUBDIR"

BIN_DIR=/image/bin
mkdir "$BIN_DIR"

cd "$NDEV_DIR"/commands || exit 1

echo Building ndau...
VERSION=$(git describe --long --tags --match="v*")
echo "  VERSION=$VERSION"
VERSION_PKG="$NDEV_SUBDIR/ndau/pkg/version"
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndaunode
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndauapi
mv ndaunode "$BIN_DIR"
mv ndauapi "$BIN_DIR"

echo Building tools...
go build ./cmd/generate
go build ./cmd/keytool
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndau
mv generate "$BIN_DIR"
mv keytool "$BIN_DIR"
mv ndau "$BIN_DIR"

echo Building procmon...
go build ./cmd/procmon
mv procmon "$BIN_DIR"

if [ -n "$RUN_UNIT_TESTS" ]; then
    echo "Running unit tests..."
    export CGO_ENABLED=0
    go test ./...
    # unfortunately, we no longer have the ability to test all oneiro
    # dependencies in any straightforward way.
    #   go test all
    # seems like a plausible approach, but it really tests _everything_,
    # and doesn't offer any subfilters. We don't want to waste time testing,
    # for example, the whole postgres suite.
    # we also can't just iterate through the output of
    #   go list -f '{{.Dir}}' -m github.com/oneiro-ndev/...
    # and run
    #   go test ./...
    # in each produced directory, because those dependencies can _only_ be
    # dependencies; they don't show up as packages which can be compiled
    # independently.
    #
    # The real solution at this point is to have each of our dependencies
    # test itself using its own circle job. That's a task for the future.
fi
