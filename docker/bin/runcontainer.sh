#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

CONTAINER=$1

if [ -z "$CONTAINER" ]; then
    CONTAINER=ndaucontainer
    echo "No container specified; using default: $CONTAINER"
fi

# Stop the container if it's running.  We can't run or restart it otherwise.
"$SCRIPT_DIR"/stopcontainer.sh "$CONTAINER"

if [ -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    echo Silencing warning about Transparent Huge Pages when redis-server runs...
    docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
           sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/enabled'
    docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
           sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/defrag'

    echo "Running ndauimage as $CONTAINER..."
    # Using --sysctl silences a warning about TCP backlog when redis runs.
    docker run -d \
           -p 26660-26661:26660-26661 \
           -p 26670-26671:26670-26671 \
           --sysctl net.core.somaxconn=511 \
           --name="$CONTAINER" \
           ndauimage 
else
    echo "Restarting $CONTAINER..."
    docker restart "$CONTAINER"
fi
echo done
