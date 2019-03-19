#!/bin/bash

CONTAINER="$1"

if [ -z "$CONTAINER" ]; then
    CONTAINER=demonet-0
    echo "No container specified; using default: $CONTAINER"
fi

# This starts a shell inside the ndau image.
docker exec -it "$CONTAINER" /bin/sh
