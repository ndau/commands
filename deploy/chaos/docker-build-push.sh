#!/bin/bash

# Check chaosnode version on ECR
# Compare this container version with ECR. Fail build if version already exists.
# Look for this sha on ecr
sha=$(git rev-parse --short "$CIRCLE_SHA1")
sha_check=$(aws ecr describe-images --repository-name chaosnode | jq ".imageDetails[].imageTags[]? | select (. == \"${sha}\")")
if [ ! -z "$sha_check" ]; then
  echo "Chaosnode container version ${sha} already exists." >&2
fi

# Build chaosnode
docker build -t -f ./chaosnode.docker /commands

# Push chaosnode
if [ "${CIRCLE_BRANCH}" == "master" ]; then
  # Do not upload commit hash that already exists.
  # Should never really happen as long as master is not tampered with.
  sha=$(git rev-parse --short "$CIRCLE_SHA1")
  sha_check=$(aws ecr describe-images --repository-name chaosnode | jq ".imageDetails[].imageTags[]? | select (. == \"${sha}\")")
  if [ ! -z "$sha_check" ]; then
    echo "Chaosnode container hash ${sha} already exists. Will not push." >&2
  else
    commit_tag="${ECR_ENDPOINT}/chaosnode:${sha}"
    latest_tag="${ECR_ENDPOINT}/chaosnode:latest"

    docker tag chaosnode $commit_tag
    docker tag chaosnode $latest_tag

    docker push $commit_tag
    docker push $latest_tag

    echo "Pushed chaosnode container hash ${sha}, and latest." >&2
  fi
fi
