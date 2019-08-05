#!/bin/bash

# This frees up space from old and unused docker images.
dead_containers=$(docker ps -q -f 'status=exited')
if [ -n "$dead_containers" ]; then
    docker rm "$dead_containers"
fi
yes | docker image prune
