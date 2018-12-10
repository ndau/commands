#!/bin/bash

# Check chaosnode version on ECR
# Compare this container version with ECR. Fail build if version already exists.
# Look for this sha on ecr
sha_check=$(aws ecr describe-images --repository-name chaosnode | jq ".imageDetails[].imageTags[]? | select (. == \"${SHA}\")")
if [ ! -z "$sha_check" ]; then
  echo "Chaosnode container version ${SHA} already exists." >&2
fi

# Build chaosnode
echo "Building chaosnode"
docker build -t chaosnode -f /commands/deploy/chaos/chaosnode.docker /commands/

# Push chaosnode
if [ "${CIRCLE_BRANCH}" == "josh/4-fix-ecr-push" ]; then
  # Do not upload commit hash that already exists.
  # Should never really happen as long as master is not tampered with.
  sha_check=$(aws ecr describe-images --repository-name chaosnode | jq ".imageDetails[].imageTags[]? | select (. == \"${SHA}\")")
  if [ ! -z "$sha_check" ]; then
    echo "Chaosnode container hash ${SHA} already exists. Will not push." >&2
  else
    commit_tag="${ECR_ENDPOINT}/chaosnode:${SHA}"
    latest_tag="${ECR_ENDPOINT}/chaosnode:latest"

    docker tag chaosnode $commit_tag
    docker tag chaosnode $latest_tag

    docker push $commit_tag
    docker push $latest_tag

    echo "Pushed chaosnode container hash ${SHA}, and latest." >&2
  fi
fi
