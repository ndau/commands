#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Set and export an environment variable for the network to use.
export NDAU_NETWORK="$1"

# This script does not support connecting to a localnet.  See run4.sh for an example of that.
if [ "$NDAU_NETWORK" != devnet ] && \
   [ "$NDAU_NETWORK" != testnet ] && \
   [ "$NDAU_NETWORK" != mainnet ]; then
    echo "Usage:"
    echo "  ./add_node_to_network [devnet|testnet|mainnet]"
    exit 1
fi

# This node is not one of the initial validators, so there's no node-identity.tgz passed in.
../bin/runcontainer.sh "$NDAU_NETWORK-test" 26666 26676 3036
