#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

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
