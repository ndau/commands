#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Start with fresh ndau/chaos and tendermint config files.
echo SETUP: Ensuring fresh configs...
rm -rf "$REDIS_CHAOS_DATA_DIR"
rm -rf "$REDIS_NDAU_DATA_DIR"
rm -rf "$NOMS_CHAOS_DATA_DIR"
rm -rf "$NOMS_NDAU_DATA_DIR"
rm -rf "$NODE_DATA_DIR"
rm -rf "$TENDERMINT_CHAOS_DATA_DIR"
rm -rf "$TENDERMINT_NDAU_DATA_DIR"

# noms
mkdir -p "$ATTICLABS_DIR"
cd "$ATTICLABS_DIR"
if [ -d "noms" ]; then
    cd noms
    ORIGIN_URL=$(git config --get remote.origin.url)
    if [ "$ORIGIN_URL" == "$NOMS_REPO" ]; then
        echo SETUP: Updating noms...
        git checkout master
        git pull origin master
    else
        echo SETUP: Replacing unsupported noms repo...
        cd ..
        rm -rf noms
        git clone "$NOMS_REPO"
        cd noms
    fi
else
    echo SETUP: Cloning noms...
    git clone "$NOMS_REPO"
    cd noms
fi
echo SETUP: Checking out noms "$NOMS_SHA"...
git checkout "$NOMS_SHA"

# tendermint
echo SETUP: Getting dep...
go get -u github.com/golang/dep/...
mkdir -p "$TM_DIR"
cd "$TM_DIR"
if [ -d "tendermint" ]; then
    echo SETUP: Updating tendermint...
    cd tendermint
    git checkout -- Gopkg.lock
    git checkout master
    git pull origin master
else
    echo SETUP: Cloning tendermint...
    git clone "$TENDERMINT_REPO"
    cd tendermint
fi
echo SETUP: Checking out tendermint "$TENDERMINT_VER"...
git checkout "$TENDERMINT_VER"
echo SETUP: Ensuring tendermint dependencies...
"$GO_DIR"/bin/dep ensure

update_repo() {
    repo="$1"
    cd "$NDEV_DIR"
    if [ -d "$repo" ]; then
        echo SETUP: Updating "$repo"...
        cd "$repo"
        git pull origin "$("$CMDBIN_DIR"/branch.sh)"
    else
        echo SETUP: Cloning "$repo"...
        git clone git@github.com:oneiro-ndev/"$repo".git
    fi
}

# ndev repos
mkdir -p "$NDEV_DIR"
update_repo commands
update_repo chaos
update_repo ndau
update_repo chaos_genesis

# utilities
cd "$NDEV_DIR"/commands
echo "SETUP: Running commands' dep ensure..."
"$GO_DIR"/bin/dep ensure

# Build everything.
echo SETUP: Building...
"$CMDBIN_DIR"/build.sh

# Test everything.
echo SETUP: Testing...
"$CMDBIN_DIR"/test.sh

# Configure everything.
echo SETUP: Configuring...
"$CMDBIN_DIR"/conf.sh

echo SETUP: Setup complete
