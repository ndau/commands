#!/bin/bash
# This script will build a base image for use in circle-ci.

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build this image
docker build --platform linux/amd64 -t circle-ci . -f "$DIR"/circle-ci.docker

# get the version label from the docker image (Docker moved it from ContainerConfig to Config)
VERSION=$(docker inspect circle-ci | jq -jr '.[0].Config.Labels["org.opencontainers.image.version"]')
echo "Version is $VERSION"
# make a version tag
TAG=578681496768.dkr.ecr.us-east-1.amazonaws.com/circle-ci:$VERSION
echo "Tag is $TAG"
# tag the image we built and push it to ECR
docker tag circle-ci:latest "$TAG"
#eval $(aws ecr get-login --no-include-email)
aws ecr get-login-password | docker login -u AWS --password-stdin "https://$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$(aws configure get region).amazonaws.com"
docker push "$TAG"
