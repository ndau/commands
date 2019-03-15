#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

# Stop the container if it's running.  We can't remove it until it's stopped.
"$SCRIPT_DIR"/stopcontainer.sh

echo Removing ndaucontainer...
docker container rm ndaucontainer 2>/dev/null
echo done
