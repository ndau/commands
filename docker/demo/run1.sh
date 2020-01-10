#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Local ip for P2P and RPC+NDAUAPI with first 4 digits of zero-based port numbers per node.
IP=$(./get_ip.sh)
P2P="$IP":2666
RPC=http://"$IP":2667

SNAPSHOT=$(./get_snapshot.sh)
IDENTITY=../ndau-snapshots/node-identity-1.tgz

../bin/runcontainer.py \
    localnet localnet-1 26661 26671 3031 \
    "$IDENTITY" \
    "$SNAPSHOT" \
    "${P2P}0" \
    "${RPC}0"
