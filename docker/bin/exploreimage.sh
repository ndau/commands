#!/bin/bash

image="$1"
if [ -z "$image" ]; then
    image=ndauimage:latest
fi

# This starts a shell inside the ndau image.
docker run --rm -it "$image" /bin/sh
