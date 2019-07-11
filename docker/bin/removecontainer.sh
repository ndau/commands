#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage:"
    echo "  ./removecontainer.sh CONTAINER"
    exit 1
fi

# Stop the container if it's running.  We can't remove it until it's stopped.
"$SCRIPT_DIR"/stopcontainer.sh "$CONTAINER"

echo "Removing $CONTAINER..."
docker container rm "$CONTAINER" 2>/dev/null
echo "done"
