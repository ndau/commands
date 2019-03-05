#!/bin/bash

echo "Starting $0"

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# grab TM version from Docker file
CONTAINER_VERSION=$(grep "org.opencontainers.image.version" "$DIR"/tendermint.docker | sed "s/.* \([v0-9].*\)/\1/")

TM_CONTAINER_VERSION=$(grep -e 'TM_VERSION_TAG' /commands/deploy/tendermint/tendermint.docker -m 1 | cut -f3 -d ' ')
version_check=$(aws ecr describe-images --repository-name tendermint | jq ".imageDetails[].imageTags[]? | select (. == \"${TM_CONTAINER_VERSION}\")")

# only push if we have a different version.
if [ ! -z "$version_check" ]; then
  echo "Tendermint container version ${TM_CONTAINER_VERSION} already exists. Will not push." >&2
  exit 0
fi

docker build -t "${ECR_ENDPOINT}/tendermint:${TM_CONTAINER_VERSION}" -f /commands/deploy/tendermint/tendermint.docker /commands
docker push "${ECR_ENDPOINT}/tendermint:${TM_CONTAINER_VERSION}"
echo "Pushed Tendermint container version ${TM_CONTAINER_VERSION}." >&2
