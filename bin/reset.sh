#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Kill everything before we wipe the pid files.
"$CMDBIN_DIR"/kill.sh

# Remove temp files.
rm -f "$CMDBIN_DIR"/*.log
rm -f "$CMDBIN_DIR"/*.pid

# Reset redis.
rm -rf "$REDIS_CHAOS_DATA_DIR"
rm -rf "$REDIS_NDAU_DATA_DIR"

# Reset noms.
rm -rf "$NOMS_CHAOS_DATA_DIR"
rm -rf "$NOMS_NDAU_DATA_DIR"

# Reset node data, such as the flag file, genesis, accounts, etc.
# We have to wipe all of that stuff since we wiped noms above and it has to get reimported.
rm -rf "$NODE_DATA_DIR"

# Reset tendermint.
cd "$TENDERMINT_DIR" || exit 1
./tendermint unsafe_reset_all --home "$TENDERMINT_CHAOS_DATA_DIR"
./tendermint unsafe_reset_all --home "$TENDERMINT_NDAU_DATA_DIR"
rm -rf "$TENDERMINT_CHAOS_DATA_DIR"
rm -rf "$TENDERMINT_NDAU_DATA_DIR"

# Reconfigure tendermint since we deleted its home directories.
"$CMDBIN_DIR"/conf.sh
