#!/bin/bash

# Dumps logs periodically.  The dump frequency is chosen to be longer than we expect a
# well-behaved node to take when starting up.  Useful in Circle CI jobs when nodes hang
# forever and are finally killed by Circle, which offers no clues as to what caused the hang.

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage:"
    echo "  ./watchcontainer.sh CONTAINER"
    exit 1
fi

echo "Watching $CONTAINER..."

# Don't watch forever.  If we've had to log a few times already, chances are we've got all the
# information we need.  This prevents callers from having to kill this script when backgrounded.
for i in {1..10}; do
    # Sleep first so we don't log anything unless we're likely in a hang situation.
    sleep 30

    echo "Still waiting for $CONTAINER"

    echo "Container logs..."
    docker container logs "$CONTAINER" 2>/dev/null | sed -e 's/^/> /'

    echo
    echo "Log files..."
    logfiles=$(docker exec "$CONTAINER" ls /image/logs)
    for logfile in ${logfiles[@]}; do
        echo "$logfile:"
        # Usually any problems are apparent at the end of the log files.
        output=$(docker exec "$CONTAINER" tail -100 "/image/logs/$logfile" | sed -e 's/^/> /')
        echo "$output"
    done
    if [ -z "$output" ]; then
        echo "(none)"
    fi
    echo
done
