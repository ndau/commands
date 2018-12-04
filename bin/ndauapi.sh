#!/bin/bash

SETUP_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
source $SETUP_DIR/env.sh

NDAUAPI_NDAU_RPC_URL=http://localhost:$TM_NDAU_RPC_PORT \
NDAUAPI_CHAOS_RPC_URL=http://localhost:$TM_CHAOS_RPC_PORT \
$NDAU_DIR/ndauapi
