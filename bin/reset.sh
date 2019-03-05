#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Kill everything before we wipe the pid files.
"$CMDBIN_DIR"/kill.sh

# Remove temp files.
rm -f "$CMDBIN_DIR"/*.log
rm -f "$CMDBIN_DIR"/*.pid

# Optionally change the number of nodes in the localnet setup.
# Do this after the above commands, so they can use the old node count.
# Do this before the steps after, so we don't possibly leave localnet in a half set up state.
node_count="$1"
if [[ ! -z "$node_count" ]]; then
    if [[ ! "$node_count" =~ ^[0-9]+$ ]]; then
        echo Node count must be a positive integer
        exit 1
    fi
    if [ "$node_count" -lt 1 ] || [ "$node_count" -gt "$MAX_NODE_COUNT" ]; then
        echo Node count must be in [1, "$MAX_NODE_COUNT"]
        exit 1
    fi

    echo "$node_count" > "$NODE_COUNT_FILE"
fi

# Reset all blockchain data.
rm -rf "$ROOT_DATA_DIR"
mkdir -p "$ROOT_DATA_DIR"

# Reconfigure everything since we deleted all the home/data directories.
"$CMDBIN_DIR"/conf.sh --needs-update
