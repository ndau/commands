#!/bin/bash

echo "Starting $0"

# Only run on master

version_check=$(aws ecr describe-images --repository-name noms | jq ".imageDetails[].imageTags[]? | select (. == \"$NOMS_CONTAINER_VERSION\")")
# only push if we have a different version.
if [ ! -z "$version_check" ]; then
  echo "Noms container version $NOMS_CONTAINER_VERSION already exists. Will not push." >&2
  exit 0
fi

docker build -t "$ECR_ENDPOINT/noms:$NOMS_CONTAINER_VERSION" -f /commands/deploy/noms/noms.docker /commands
docker push "$ECR_ENDPOINT/noms:$NOMS_CONTAINER_VERSION"
echo "Pushed Noms container version $NOMS_CONTAINER_VERSION." >&2

