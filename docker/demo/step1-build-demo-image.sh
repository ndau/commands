#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "$SCRIPT_DIR"/../bin || exit 1

./buildimage.sh "$SCRIPT_DIR"/genesis-demo.toml
