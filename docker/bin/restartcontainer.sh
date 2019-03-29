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

# Sleep a bit to give it a chance to remove the 'running' file as it starts up.
# This prevents the wait loop below from exiting early if it sees an old copy of that file.
sleep 1

echo "Waiting for the node to fully spin up..."
until docker exec "$CONTAINER" test -f /image/running 2>/dev/null
do
    :
done

echo done
