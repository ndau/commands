#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

BIN_DIR="$SCRIPT_DIR"/../bin

"$BIN_DIR"/buildimage.sh "$SCRIPT_DIR/genesis-demo.toml"
