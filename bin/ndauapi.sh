#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# We use the ports of the 0'th node, even in a multi-node localnet.
chaos_rpc_port="$TM_RPC_PORT"
ndau_rpc_port=$(expr "$TM_RPC_PORT" + 1)

NDAUAPI_NDAU_RPC_URL=http://localhost:"$ndau_rpc_port" \
NDAUAPI_CHAOS_RPC_URL=http://localhost:"$chaos_rpc_port" \
"$COMMANDS_DIR"/ndauapi
