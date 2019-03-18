#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

GENESIS_FILE=$1

if [ -z "$GENESIS_FILE" ]; then
    echo Usage:
    echo "  ./buildimage.sh PATH_TO_GENESIS_TOML"
    exit 1
fi

if [ ! -e "$GENESIS_FILE" ]; then
    echo "Cannot find genesis.toml file $GENESIS_FILE"
    exit 1
fi

# Remove the container if it exists.  We don't want it around since it's based of an old image.
"$SCRIPT_DIR"/removecontainer.sh "$CONTAINER"

echo Removing ndauimage...
docker image rm ndauimage 2>/dev/null
echo done

echo Building ndauimage...
docker build "$SCRIPT_DIR"/../image --tag=ndauimage
echo done
