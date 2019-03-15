#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

# Stop the container if it's running.  We can't run or restart it otherwise.
"$SCRIPT_DIR"/stopcontainer.sh

if [ -z "$(docker container ls -a -q -f name=ndaucontainer)" ]; then
    echo Running ndauimage as ndaucontainer...
    docker run -d \
           -p 26660-26661:26660-26661 \
           -p 26670-26671:26670-26671 \
           --name=ndaucontainer \
           ndauimage 
else
    echo Restarting ndaucontainer...
    docker restart ndaucontainer
fi
echo done
