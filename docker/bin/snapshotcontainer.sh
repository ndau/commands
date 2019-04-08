#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage:"
    echo "  ./snapshotcontainer.sh CONTAINER"
    exit 1
fi

if [ -z "$(docker container ls -q -f name=$CONTAINER)" ]; then
    echo "Container does not exist or is not running: $CONTAINER"
    echo "Use runcontainer.sh or restartcontainer.sh to start the container first"
    exit 1
fi

echo "Generating $CONTAINER snapshot..."
docker exec "$CONTAINER" /image/docker-snapshot.sh

echo "Waiting for snapshot..."
until docker exec "$CONTAINER" test -f /image/snapshot_result 2>/dev/null
do
    :
done

# Get the contents of the file.  We could have folded this into the test above, but this way helps
# ensure we don't read partial data as the file is being written to inside the container.
SNAPSHOT_RESULT=$(docker exec "$CONTAINER" cat /image/snapshot_result)

if [[ "$SNAPSHOT_RESULT" == "ERROR:"* ]]; then
    echo "SNAPSHOT_RESULT"
    exit 1
fi

OUT_FILE="$SCRIPT_DIR/$SNAPSHOT_RESULT"
docker cp "$CONTAINER:/image/$SNAPSHOT_RESULT" "$OUT_FILE"

echo "The snapshot has been generated and copied out of the container here:"
echo "  $OUT_FILE"
