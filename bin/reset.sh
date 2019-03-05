#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Optionally change the number of nodes in the localnet setup.
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

    export NODE_COUNT="$node_count"
    export HIGH_NODE_NUM=$((NODE_COUNT - 1))
fi

# Kill everything before we wipe the pid files.
"$CMDBIN_DIR"/kill.sh

# Remove temp files.
rm -f "$CMDBIN_DIR"/*.log
rm -f "$CMDBIN_DIR"/*.pid

# Reset all blockchain data.
rm -rf "$ROOT_DATA_DIR"
mkdir -p "$ROOT_DATA_DIR"

# Reconfigure everything since we deleted all the home/data directories.
"$CMDBIN_DIR"/conf.sh --needs-update
