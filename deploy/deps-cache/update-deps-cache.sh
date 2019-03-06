#!/bin/bash
# This script will build a base image that helps speed up circle-ci's build process.

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Check to see if machine_user_key is two directories up (location for everything else).
if [ ! -f "$DIR/../../machine_user_key" ]; then
    >&2 echo "machine_user_key not in project root."
    exit 1
fi

# get the version label from the docker file
CONTAINER_VERSION=$(grep "org.opencontainers.image.version.*" "$DIR"/deps-cache.docker | sed "s/.* \([v0-9].*\)/\1/")

version_check=$(aws ecr describe-images --repository-name deps-cache | jq ".imageDetails[].imageTags[]? | select (. == \"$CONTAINER_VERSION\")")
# only push if we have a different version.
if [ ! -z "$version_check" ]; then
  echo "deps-cache container version $CONTAINER_VERSION already exists. Will not push." >&2
  exit 0
fi
# make a version tag
TAG=578681496768.dkr.ecr.us-east-1.amazonaws.com/deps-cache:$CONTAINER_VERSION

# Build this image
docker build -t "$TAG" /commands -f "$DIR"/deps-cache.docker

# push the image to ECR
docker push "$TAG"
