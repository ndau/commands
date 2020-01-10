#!/bin/bash
# This script will build a base image for use in circle-ci.

set -e

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build this image
cd "$DIR"
docker build -t circle-ci . -f circle-ci.docker

# get the version label from the docker image
VERSION=$(docker inspect circle-ci | jq -jr '.[0].ContainerConfig.Labels["org.opencontainers.image.version"]')

# make a version tag
TAG=578681496768.dkr.ecr.us-east-1.amazonaws.com/circle-ci:$VERSION

# tag the image we built and push it to ECR
docker tag circle-ci "$TAG"
eval $(aws ecr get-login --no-include-email)
docker push "$TAG"
