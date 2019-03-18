SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

TENDERMINT_VER=v0.30.1

ATTICLABS_DIR="$GOPATH"/src/github.com/attic-labs
NDEV_SUBDIR=github.com/oneiro-ndev
NDEV_DIR="$GOPATH"/src/"$NDEV_SUBDIR"
TM_DIR="$GOPATH"/src/github.com/tendermint

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

echo Patching tendermint...
patch -i "$SCRIPT_DIR"/Gopkg.toml.patch Gopkg.toml
patch -i "$SCRIPT_DIR"/root.go.patch cmd/tendermint/commands/root.go

echo Getting ndev repositories...
mkdir -p "$NDEV_DIR"
cd "$NDEV_DIR" || exit 1
git clone git@github.com:oneiro-ndev/commands.git
git clone git@github.com:oneiro-ndev/ndau.git

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

echo Building tendermint...
cd "$TM_DIR"/tendermint || exit 1
go build ./cmd/tendermint

echo Building chaos...
cd "$NDEV_DIR"/commands || exit 1
go build ./cmd/chaos
go build ./cmd/chaosnode

echo Building ndau...
cd "$NDEV_DIR"/ndau || exit 1
VERSION=$(git describe --long --tags --match="v*")
VERSION_PKG="$NDEV_SUBDIR/commands/vendor/$NDEV_SUBDIR/ndau/pkg/version"
cd "$NDEV_DIR"/commands || exit 1
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndau
go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./cmd/ndaunode
go build ./cmd/ndauapi

echo Setup complete
