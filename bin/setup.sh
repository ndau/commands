#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Make sure we have the deploy file.
# If this exits with error, see README.md for how to get the deploy file.
DEPLOY_FILE=$(pwd)/machine_user_key
echo SETUP: Ensuring "$DEPLOY_FILE" exists...
stat "$DEPLOY_FILE" >/dev/null

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
    echo SETUP: Updating noms...
    cd noms
    git pull origin "$("$CMDBIN_DIR"/branch.sh)"
else
    echo SETUP: Cloning noms...
    git clone https://github.com/oneiro-ndev/noms.git
fi

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
    git pull origin "$("$CMDBIN_DIR"/branch.sh)"
else
    echo SETUP: Cloning tendermint...
    git clone https://github.com/tendermint/tendermint.git
fi
echo SETUP: Checking out tendermint "$TENDERMINT_VER"...
git checkout "$TENDERMINT_VER"
echo SETUP: Ensuring tendermint dependencies...
"$GO_DIR"/bin/dep ensure

# ndev repos
mkdir -p "$NDEV_DIR"
cd "$NDEV_DIR"
if [ -d "chaos" ]; then
    echo SETUP: Updating chaos...
    cd chaos
    git pull origin "$("$CMDBIN_DIR"/branch.sh)"
else
    echo SETUP: Cloning chaos...
    git clone git@github.com:oneiro-ndev/chaos.git
fi
cd "$NDEV_DIR"
if [ -d "ndau" ]; then
    echo SETUP: Updating ndau...
    cd ndau
    git pull origin "$("$CMDBIN_DIR"/branch.sh)"
else
    echo SETUP: Cloning ndau...
    git clone git@github.com:oneiro-ndev/ndau.git
fi
cd "$NDEV_DIR"
if [ -d "chaos_genesis" ]; then
    echo SETUP: Updating chaos_genesis...
    cd chaos_genesis
    git pull origin "$("$CMDBIN_DIR"/branch.sh)"
else
    echo SETUP: Cloning chaos_genesis...
    git clone git@github.com:oneiro-ndev/chaos_genesis.git
fi

# utilities
cd "$NDEV_DIR"/commands
if [ "$DEPLOY_FILE" != "$(pwd)/machine_user_key" ]; then
    cp "$DEPLOY_FILE" .
fi
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
