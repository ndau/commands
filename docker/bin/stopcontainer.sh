#!/bin/bash

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage:"
    echo "  ./stopcontainer.sh CONTAINER"
    exit 1
fi

echo Stopping "$CONTAINER"...
docker container stop "$CONTAINER" 2>/dev/null
echo done
