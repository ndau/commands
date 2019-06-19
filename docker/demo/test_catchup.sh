#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Test catchup on testnet from block 1.
# We use the mainnet genesis snapshot for this since testnet is deployed with mainnet data.
NETWORK="testnet"
SNAPSHOT="snapshot-mainnet-1"

# Use the latest built ndauimage from the local docker environment.
export USE_LOCAL_IMAGE=1

# When testing new features for catchup-compatibility, we set this env var at node-deploy-time.
# Set this to a height in the future.
export NEW_FEATURE_HEIGHT=15000

../bin/runcontainer.sh "$NETWORK" "$NETWORK-test" 26666 26676 3036 "" "$SNAPSHOT"
