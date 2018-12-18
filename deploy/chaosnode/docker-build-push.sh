#!/bin/bash

echo "Starting $0"

# Check chaosnode version on ECR

# Look for container tagged with this SHA on ECR.
sha_check=$(aws ecr describe-images --repository-name chaosnode | jq ".imageDetails[].imageTags[]? | select (. == \"$SHA\")")

if [ ! -z "$sha_check" ]; then
  echo "Chaosnode container version $SHA already exists. Will not build. Will not push." >&2
  exit 0
fi

# Build chaosnode
docker build -t chaosnode -f /commands/deploy/chaosnode/chaosnode.docker /commands/

# compose tags for ecr
commit_tag="${ECR_ENDPOINT}/chaosnode:$SHA"
latest_tag="${ECR_ENDPOINT}/chaosnode:latest"

docker tag chaosnode $commit_tag
docker tag chaosnode $latest_tag

docker push $commit_tag
docker push $latest_tag

echo "Pushed chaosnode container hash $SHA, and latest." >&2

