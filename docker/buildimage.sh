#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

# Remove the container if it exists.  We don't want it around since it's based of an old image.
"$SCRIPT_DIR"/removecontainer.sh

echo Removing ndauimage...
docker image rm ndauimage 2>/dev/null
echo done

echo Building ndauimage...
docker build "$SCRIPT_DIR"/image --tag=ndauimage
echo done
