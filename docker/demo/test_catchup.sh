#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Test catchup on mainnet from block 1.
NETWORK="mainnet"
SNAPSHOT="snapshot-mainnet-1"

# Use the latest built ndauimage from the local docker environment.
export USE_LOCAL_IMAGE=1

../bin/runcontainer.sh "$NETWORK" "$NETWORK-test" 26666 26676 3036 "" "$SNAPSHOT"
