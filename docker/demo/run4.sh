#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Local ip for P2P and RPC+NDAUAPI with first 4 digits of zero-based port numbers per node.
IP=$(./get_ip.sh)
P2P="$IP":2666
RPC=http://"$IP":2667

SNAPSHOT=$(./get_snapshot.sh)
IDENTITY=../ndau-snapshots/node-identity-4.tgz

../bin/runcontainer.sh \
    localnet-4 26664 26674 3034 \
    "$IDENTITY" \
    "$SNAPSHOT" \
    "${P2P}0,${P2P}1,${P2P}2,${P2P}3" \
    "${RPC}0,${RPC}1,${RPC}2,${RPC}3"
