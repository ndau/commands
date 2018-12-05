#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
SETUP_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
# shellcheck disable=SC1090
source "$SETUP_DIR"/env.sh

# Ensure the go directory is where we expect it.
echo SETUP: Ensuring "$GO_DIR" exists...
if [[ $GO_DIR == *":"* ]]; then
    echo Multiple Go paths not supported
    exit 1
fi
mkdir -p "$GO_DIR"

# Make sure we have the deploy file.
# If this exits with error, see README.md for how to get the deploy file.
DEPLOY_FILE=$SETUP_DIR/github_chaos_deploy
echo SETUP: Ensuring "$DEPLOY_FILE" exists...
# STAT=`stat "$DEPLOY_FILE"`

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
    git pull origin "$("$SETUP_DIR"/branch.sh)"
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
    git pull origin "$("$SETUP_DIR"/branch.sh)"
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
    git pull origin "$("$SETUP_DIR"/branch.sh)"
else
    echo SETUP: Cloning chaos...
    git clone git@github.com:oneiro-ndev/chaos.git
fi
cd "$NDEV_DIR"
if [ -d "ndau" ]; then
    echo SETUP: Updating ndau...
    cd ndau
    git pull origin "$("$SETUP_DIR"/branch.sh)"
else
    echo SETUP: Cloning ndau...
    git clone git@github.com:oneiro-ndev/ndau.git
fi
cd "$NDEV_DIR"
if [ -d "chaos_genesis" ]; then
    echo SETUP: Updating chaos_genesis...
    cd chaos_genesis
    git pull origin "$("$SETUP_DIR"/branch.sh)"
else
    echo SETUP: Cloning chaos_genesis...
    git clone git@github.com:oneiro-ndev/chaos_genesis.git
fi

# chaos tools
cd "$NDEV_DIR"/chaos
cp "$DEPLOY_FILE" .
echo SETUP: Running chaos glide install...
glide install

# ndau tools
cd "$NDEV_DIR"/ndau
cp "$DEPLOY_FILE" .
echo SETUP: Running ndau glide install...
glide install

# chaos_genesis tools
cd "$NDEV_DIR"/chaos_genesis
echo SETUP: Running chaos_genesis glide install...
glide install

# Build everything.
echo SETUP: Building...
"$SETUP_DIR"/build.sh

# Test everything.
echo SETUP: Testing...
"$SETUP_DIR"/test.sh

# Configure everything.
echo SETUP: Configuring...
"$SETUP_DIR"/conf.sh

echo SETUP: Setup complete
