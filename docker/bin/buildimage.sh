#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

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
echo done

echo "Building $NDAU_IMAGE_NAME..."
docker build \
       --build-arg SSH_PRIVATE_KEY="$SSH_PRIVATE_KEY" \
       "$DOCKER_DIR"/image \
       --tag="$NDAU_IMAGE_NAME"
echo done

# Must have latest ndau tool built locally to get the version in order to save/upload the image.
# If you don't have it, it's fine.  We'll just skip the option of saving/uploading to S3.
NDAU_TOOL="$SCRIPT_DIR/../../ndau"
if [ -f "$NDAU_TOOL" ]; then
    echo "Saving local copy of $NDAU_IMAGE_NAME..."

    VERSION=$($NDAU_TOOL version)

    NDAU_IMAGES_SUBDIR=ndau-images
    NDAU_IMAGES_DIR="$DOCKER_DIR/$NDAU_IMAGES_SUBDIR"
    mkdir -p "$NDAU_IMAGES_DIR"

    # Save the docker image.
    IMAGE_NAME="$NDAU_IMAGE_NAME-$VERSION"
    IMAGE_PATH="$NDAU_IMAGES_DIR/$IMAGE_NAME.docker"
    docker save -o "$IMAGE_PATH" "$NDAU_IMAGE_NAME"
    gzip -f "$IMAGE_PATH"
    IMAGE_PATH="$IMAGE_PATH.gz"

    # Save the version file.
    VERSION_FILE=latest.txt
    VERSION_PATH="$NDAU_IMAGES_DIR/$VERSION_FILE"
    echo "$IMAGE_NAME" > "$VERSION_PATH"

    # These can be used for uploading the snapshot to S3.
    S3_DIR_URI="s3://$NDAU_IMAGES_SUBDIR"
    UPLOAD_IMAGE_CMD="aws s3 cp $IMAGE_PATH $S3_DIR_URI/$IMAGE_NAME.docker.gz"
    UPLOAD_VERSION_CMD="aws s3 cp $VERSION_PATH $S3_DIR_URI/$VERSION_FILE"

    echo
    echo "IMAGE CREATED: $IMAGE_PATH"
    echo "VERSION CREATED: $VERSION_PATH"
    echo
    echo "Optional next steps:"
    echo "  1. Upload the image to S3 using:"
    echo "       $UPLOAD_IMAGE_CMD"
    echo "  2. If the image was uploaded, use this to mark it as the latest if desired:"
    echo "       $UPLOAD_VERSION_CMD"
    echo
fi
