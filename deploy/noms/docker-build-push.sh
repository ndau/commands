#!/bin/bash

echo "Starting $0"

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CONTAINER_VERSION=$(grep "org.opencontainers.image.version.*" "$DIR"/noms.docker |  sed "s/.* \([v0-9].*\)/\1/")

version_check=$(aws ecr describe-images --repository-name noms | jq ".imageDetails[].imageTags[]? | select (. == \"$CONTAINER_VERSION\")")
# only push if we have a different version.
if [ ! -z "$version_check" ]; then
  echo "Noms container version $CONTAINER_VERSION already exists. Will not push." >&2
  exit 0
fi

docker build -t "$ECR_ENDPOINT/noms:$CONTAINER_VERSION" -f "$DIR"/noms.docker /commands
docker push "$ECR_ENDPOINT/noms:$CONTAINER_VERSION"
echo "Pushed Noms container version $CONTAINER_VERSION." >&2

