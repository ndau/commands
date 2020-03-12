#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

COMMANDS_BRANCH="$1"
if [ -z "$COMMANDS_BRANCH" ]; then
    COMMANDS_BRANCH=$(git symbolic-ref --short HEAD 2> /dev/null)
    if [ -z "$COMMANDS_BRANCH" ]; then
        echo "No commands branch specified"
        exit 1
    fi
fi
echo "Using commands branch/tag: $COMMANDS_BRANCH"

DOCKER_DIR="$SCRIPT_DIR/.."
IMAGE_DIR="$DOCKER_DIR/image"
COMMANDS_DIR="$DOCKER_DIR/.."
SSH_PRIVATE_KEY_FILE="$COMMANDS_DIR"/machine_user_key
if [ ! -e "$SSH_PRIVATE_KEY_FILE" ]; then
    # This file can be gotten from Oneiro's 1password account and placed in the docker directory.
    echo "Cannot find $SSH_PRIVATE_KEY_FILE needed for cloning private ndau repositories"
    exit 1
fi
SSH_PRIVATE_KEY=$(cat "$SSH_PRIVATE_KEY_FILE")

NDAU_IMAGE_NAME=ndauimage
if [ -n "$(docker container ls -a -q -f ancestor=$NDAU_IMAGE_NAME)" ]; then
    echo "-------"
    echo "WARNING: containers exist based on an old $NDAU_IMAGE_NAME; they should be removed"
    echo "-------"
fi

# update shas for cache-busting when appropriate
curl -s https://api.github.com/repos/attic-labs/noms/git/refs/heads/master |\
    jq -r .object.sha > "$IMAGE_DIR/noms_sha"
git rev-parse HEAD > "$IMAGE_DIR/commands_sha"
if [ -n "$(git status --porcelain)" ]; then
    echo "WARN: uncommitted changes"
    echo "docker image contains only committed work ($(git rev-parse --short HEAD))"
fi

# update dependencies for cache-busting when appropriate
cp "$COMMANDS_DIR"/Gopkg.* "$IMAGE_DIR"/

cd "$COMMANDS_DIR" || exit 1
SHA=$(git rev-parse --short HEAD)

echo "Building $NDAU_IMAGE_NAME..."
if ! docker build \
       --build-arg SSH_PRIVATE_KEY="$SSH_PRIVATE_KEY" \
       --build-arg COMMANDS_BRANCH="$COMMANDS_BRANCH" \
       --build-arg RUN_UNIT_TESTS="$RUN_UNIT_TESTS" \
       "$IMAGE_DIR" \
       --tag="$NDAU_IMAGE_NAME:$SHA" \
       --tag="$NDAU_IMAGE_NAME:latest"
then
    echo "Failed to build $NDAU_IMAGE_NAME"
    exit 1
fi

echo "done"
