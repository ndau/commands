#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Devnet uses two different domain names for P2P and RPC+NDAUPI.  Here they are, with the first
# four digits of the corresponding 5-digit zero-based port numbers for each of 5 starting nodes.
P2P=p2p.ndau.tech:3020
RPC=https://api.ndau.tech:3010

# Devnet uses the same snapshot as testnet.
SNAPSHOT=snapshot-testnet-47

../bin/runcontainer.sh \
    devnet-test 26665 26675 3035 \
    "${P2P}0,${P2P}1,${P2P}2,${P2P}3,${P2P}4" \
    "${RPC}0,${RPC}1,${RPC}2,${RPC}3,${RPC}4" \
    $SNAPSHOT
# This node is not one of the initial validators, so there's no node-identity.tgz passed in.
