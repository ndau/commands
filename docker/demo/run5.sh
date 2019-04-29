#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Local ip for P2P and RPC+NDAUAPI with first 4 digits of zero-based port numbers per node.
IP=$(./get_ip.sh)
P2P="$IP":2666
RPC=http://"$IP":2667

SNAPSHOT=$(./get_snapshot.sh)
IDENTITY="" # This last node demonstrates starting a node and having its identity file generated.

../bin/runcontainer.sh \
    localnet localnet-5 26665 26675 3035 \
    "$IDENTITY" \
    "$SNAPSHOT" \
    "${P2P}0,${P2P}1,${P2P}2,${P2P}3,${P2P}4" \
    "${RPC}0,${RPC}1,${RPC}2,${RPC}3,${RPC}4"
