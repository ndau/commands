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
    # It takes multiple seconds to take a snapshot, so checking once per second doesn't cause too
    # much extra wait time and it also frees up CPU for the node to consume while snapshotting.
    sleep 1
done

# Get the contents of the file.  We could have folded this into the test above, but this way helps
# ensure we don't read partial data as the file is being written to inside the container.
SNAPSHOT_RESULT=$(docker exec "$CONTAINER" cat /image/snapshot_result)

SNAPSHOT_FILE="$SCRIPT_DIR/$SNAPSHOT_RESULT"
docker cp "$CONTAINER:/image/$SNAPSHOT_RESULT" "$SNAPSHOT_FILE"

# These can be used for uploading the snapshot to S3.
S3URI="s3://ndau-snapshots/$SNAPSHOT_RESULT"
UPLOAD_CMD="aws s3 cp $SNAPSHOT_FILE $S3URI"

echo
echo "The snapshot has been generated and copied out of the container here:"
echo "  $SNAPSHOT_FILE"
echo "It can be uploaded to S3 using the following command:"
echo "  $UPLOAD_CMD"
echo
