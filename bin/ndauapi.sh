#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

NDAUAPI_NDAU_RPC_URL=http://localhost:$TM_NDAU_RPC_PORT \
NDAUAPI_CHAOS_RPC_URL=http://localhost:"$TM_CHAOS_RPC_PORT" \
"$NDAUAPI_CMD"/ndauapi
