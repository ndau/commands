#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

COMMANDS_BRANCH="$1"
if [ -z "$COMMANDS_BRANCH" ]; then
    COMMANDS_BRANCH=$(git symbolic-ref --short HEAD 2> /dev/null)
fi
echo "Using commands branch/tag: $COMMANDS_BRANCH"

DOCKER_DIR="$SCRIPT_DIR/.."
COMMANDS_DIR="$DOCKER_DIR/.."
SSH_PRIVATE_KEY_FILE="$COMMANDS_DIR"/machine_user_key
if [ ! -e "$SSH_PRIVATE_KEY_FILE" ]; then
    # This file can be gotten from Oneiro's 1password account and placed in the docker directory.
    echo "Cannot find $SSH_PRIVATE_KEY_FILE needed for cloning private oneiro-ndev repositories"
    exit 1
fi
SSH_PRIVATE_KEY=$(cat "$SSH_PRIVATE_KEY_FILE")

NDAU_IMAGE_NAME=ndauimage
if [ ! -z "$(docker container ls -a -q -f ancestor=$NDAU_IMAGE_NAME)" ]; then
    echo "-------"
    echo "WARNING: containers exist based on an old $NDAU_IMAGE_NAME; they should be removed"
    echo "-------"
fi

echo "Removing $NDAU_IMAGE_NAME..."
docker image rm "$NDAU_IMAGE_NAME" 2>/dev/null
echo "done"

echo "Building $NDAU_IMAGE_NAME..."
# Use --no-cache since we likely have new source code we want built that docker can't detect.
docker build \
       --no-cache \
       --build-arg SSH_PRIVATE_KEY="$SSH_PRIVATE_KEY" \
       --build-arg COMMANDS_BRANCH="$COMMANDS_BRANCH" \
       "$DOCKER_DIR"/image \
       --tag="$NDAU_IMAGE_NAME"
echo "done"
