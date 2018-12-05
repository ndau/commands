#!/bin/bash

# Print the branch name of the git repo in the current directory.
BRANCH=$(git symbolic-ref --short HEAD 2> /dev/null)
echo "$BRANCH"
