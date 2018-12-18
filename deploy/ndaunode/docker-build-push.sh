#!/bin/bash

echo "Starting $0"

# Push ndaunode image to ECR
# Do not upload commit hash that already exists.
# Should never really happen as long as the ECR_PUSH_BRANCH is not tampered with.
sha_check=$(aws ecr describe-images --repository-name ndaunode | jq ".imageDetails[].imageTags[]? | select (. == \"$SHA\")")
if [ ! -z "$sha_check" ]; then
  echo "Ndaunode container hash $SHA already exists. Will not push." >&2
  # If a container with this hash is built already, docker will overwrite
  # with a push. This could change the containers behavior if, for example,
  # dependencies have changed between subsequent build times.
  exit 0
fi

# Build ndaunode-build
docker build -t ndaunode-build -f /commands/deploy/ndaunode/ndaunode-build.docker /commands --build-arg VERSION=$VERSION

# Build ndaunode-run
docker build -t ndaunode -f /commands/deploy/ndaunode/ndaunode-run.docker /commands

commit_tag="${ECR_ENDPOINT}/ndaunode:$SHA"
latest_tag="${ECR_ENDPOINT}/ndaunode:latest"

docker tag ndaunode $commit_tag
docker tag ndaunode $latest_tag

docker push $commit_tag
docker push $latest_tag

echo "Pushed ndaunode with tags :$SHA, :latest." >&2
