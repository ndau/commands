#!/bin/bash

# push the ndauapi image to ECR
# Do not upload commit hash that already exists.
# Should never really happen as long as master is not tampered with.
sha_check=$(aws ecr describe-images --repository-name ndauapi | jq ".imageDetails[].imageTags[]? | select (. == \"$SHA\")")

if [ ! -z "$sha_check" ]; then
  echo "ndauapi container hash $SHA already exists. Will not push." >&2
  exit 0
fi

# Build ndauapi-build
docker build -t ndauapi-build -f /commands/deploy/ndau/ndauapi-build.docker /commands

# Build ndauapi-run
docker build -t ndauapi -f /commands/deploy/ndau/ndauapi-run.docker /commands

commit_tag="${ECR_ENDPOINT}/ndauapi:$SHA"
latest_tag="${ECR_ENDPOINT}/ndauapi:latest"

docker tag ndauapi $commit_tag
docker tag ndauapi $latest_tag

docker push $commit_tag
docker push $latest_tag

echo "Pushed ndauapi with tags :$SHA, :latest." >&2
