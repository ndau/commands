#!/bin/bash

# This frees up space from old and unused docker images.
docker rm $(docker ps -q -f 'status=exited')
docker rmi $(docker images -q -f "dangling=true")
