#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

GENESIS_TOML_FILE="$1"

if [ -z "$GENESIS_TOML_FILE" ]; then
    echo Usage:
    echo "  ./buildimage.sh GENESIS_TOML_FILE"
    exit 1
fi
if [ ! -e "$GENESIS_TOML_FILE" ]; then
    echo "Cannot find genesis toml file $GENESIS_TOML_FILE"
    exit 1
fi
GENESIS_TOML=$(cat "$GENESIS_TOML_FILE")

DOCKER_DIR="$SCRIPT_DIR/.."
SSH_PRIVATE_KEY_FILE="$DOCKER_DIR"/machine_user_key
if [ ! -e "$SSH_PRIVATE_KEY_FILE" ]; then
    # This file can be gotten from Oneiro's 1password account and placed in the docker directory.
    echo "Cannot find $SSH_PRIVATE_KEY_FILE needed for cloning private oneiro-ndev repositories"
    exit 1
fi
SSH_PRIVATE_KEY=$(cat "$SSH_PRIVATE_KEY_FILE")

if [ ! -z "$(docker container ls -a -q -f ancestor=ndauimage)" ]; then
    echo "-------"
    echo "WARNING: containers exist based on an old ndauimage; they should be removed"
    echo "-------"
fi

echo Removing ndauimage...
docker image rm ndauimage 2>/dev/null
echo done

echo Preparing image directory...
IMAGE_DIR="$DOCKER_DIR"/image
COMMANDS_DIR="$DOCKER_DIR/.."
PATCH_DIR="$COMMANDS_DIR"/deploy/tendermint
git clean -fx "$IMAGE_DIR"
cp "$PATCH_DIR"/*.patch "$IMAGE_DIR"

echo Building ndauimage...
docker build \
       --build-arg GENESIS_TOML="$GENESIS_TOML" \
       --build-arg SSH_PRIVATE_KEY="$SSH_PRIVATE_KEY" \
       --squash \
       "$DOCKER_DIR"/image \
       --tag=ndauimage
echo done
