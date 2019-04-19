#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Local ip for P2P and RPC+NDAUAPI with first 4 digits of zero-based port numbers per node.
IP=$(./get_ip.sh)
P2P="$IP":2666
RPC=http://"$IP":2667

SNAPSHOT=$(./get_snapshot.sh)
IDENTITY=../../bin/ndau-snapshots/node-identity-3.tgz

../bin/runcontainer.sh \
    localnet-3 26663 26673 3033 \
    "$SNAPSHOT" \
    "$IDENTITY" \
    "${P2P}0,${P2P}1,${P2P}2" \
    "${RPC}0,${RPC}1,${RPC}2"
