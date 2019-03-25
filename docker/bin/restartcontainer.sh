#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage:"
    echo "  ./restartcontainer.sh CONTAINER"
    exit 1
fi

if [ -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    echo "Container does not exist: $CONTAINER"
    echo "Use runcontainer.sh to run a container for the first time"
    exit 1
fi    

echo "Restarting $Container..."
docker restart "$CONTAINER"
