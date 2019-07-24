#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

ATTICLABS_DIR="$GOPATH"/src/github.com/attic-labs
NDEV_SUBDIR=github.com/oneiro-ndev
NDEV_DIR="$GOPATH/src/$NDEV_SUBDIR"
TM_DIR="$GOPATH"/src/github.com/tendermint

BIN_DIR="$SCRIPT_DIR"/bin
mkdir "$BIN_DIR"

echo Getting noms...
mkdir -p "$ATTICLABS_DIR"
cd "$ATTICLABS_DIR" || exit 1
git clone git@github.com:oneiro-ndev/noms.git

echo Getting tendermint...
mkdir -p "$TM_DIR"
cd "$TM_DIR" || exit 1
git clone https://github.com/tendermint/tendermint.git

echo Checking out tendermint "$TENDERMINT_VER"...
cd tendermint || exit 1
git fetch --prune
git checkout "$TENDERMINT_VER"

echo "Getting commands $COMMANDS_BRANCH branch..."
mkdir -p "$NDEV_DIR"
cd "$NDEV_DIR" || exit 1
git clone git@github.com:oneiro-ndev/commands.git --branch "$COMMANDS_BRANCH"

echo Installing dep...
go get -u github.com/golang/dep/...

echo Ensuring dependencies for tendermint...
cd "$TM_DIR"/tendermint || exit 1
"$GOPATH"/bin/dep ensure

echo Ensuring dependencies for commands...
cd "$NDEV_DIR"/commands || exit 1
"$GOPATH"/bin/dep ensure

echo Building noms...
cd "$ATTICLABS_DIR"/noms || exit 1
go build ./cmd/noms
mv noms "$BIN_DIR"

echo Building tendermint...
cd "$TM_DIR"/tendermint || exit 1
go build ./cmd/tendermint
mv tendermint "$BIN_DIR"

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
