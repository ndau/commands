#!/bin/bash

# clean up dead containers
dead_containers=$(docker ps -q -f 'status=exited')
if [ -n "$dead_containers" ]; then
    docker rm "$dead_containers"
fi
# clean up dangling images
docker image prune -f
# clean up old version of ndauimage
docker images ndauimage --filter=before=ndauimage:latest --format="{{.ID}}" |\
    xargs docker image rm
