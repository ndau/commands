#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

NODE0_CONTAINER=demonet-0
NODE0_CHAOS_P2P=26660
NODE0_CHAOS_RPC=26670
NODE0_NDAU_P2P=26661
NODE0_NDAU_RPC=26671

cd "$SCRIPT_DIR"/../bin || exit 1

./runcontainer.sh \
    "$NODE0_CONTAINER" \
    "$NODE0_CHAOS_P2P" \
    "$NODE0_CHAOS_RPC" \
    "$NODE0_NDAU_P2P" \
    "$NODE0_NDAU_RPC"
