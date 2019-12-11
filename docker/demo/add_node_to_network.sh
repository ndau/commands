#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Pass in the network name.  Supported networks: devnet, testnet, mainnet.
NETWORK="$1"

# This script demonstrates joining a network without specifying peers.
# It does not support connecting to a localnet.  See run5.sh for an example of that.
if [ "$NETWORK" != devnet ] && \
   [ "$NETWORK" != testnet ] && \
   [ "$NETWORK" != mainnet ]; then
    echo "Usage:"
    echo "  ./add_node_to_network [devnet|testnet|mainnet]"
    exit 1
fi

# This node is not one of the initial validators, so there's no node-identity.tgz passed in.
../bin/runcontainer.py "$NETWORK" "$NETWORK-test" 26666 26676 3036
