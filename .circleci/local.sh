#!/bin/bash
# This script runs the circle ci build on your local machine.

# errcho echos to stderr
errcho(){
    echo -e "$@" >&2
}

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$(which circleci)" ]; then
    errcho "Installing circleci..."
    curl -o /usr/local/bin/circleci https://circle-downloads.s3.amazonaws.com/releases/build_agent_wrapper/circleci
    chmod +x /usr/local/bin/circleci
fi

# set flag that gets turned to true if there are any errors
errors=false

# check for necessary files

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    errors=true
    errcho "Missing AWS_ACCESS_KEY_ID env var"
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    errors=true
    errcho "Missing AWS_SECRET_ACCESS_KEY env var"
fi

if [ -z "$AWS_DEPLOY_SECRETS_ID" ]; then
    errors=true
    errcho "Missing AWS_DEPLOY_SECRETS_ID env var"
fi

if [ -z "$AWS_DEPLOY_SECRETS_KEY" ]; then
    errors=true
    errcho "Missing AWS_DEPLOY_SECRETS_KEY env var"
fi

if [ ! -f "$DIR/../machine_user_key" ]; then
    errors=true
    errcho "Missing $DIR/../machine_user_key"
fi

if [ ! -f "$HOME/.ssh/sc-node-ec2.pem" ]; then
    errors=true
    errcho "Missing $HOME/.ssh/sc-node-ec2.pem"
fi

# exit if there's any errors
$errors && exit 1

# escape newlines removes newline characters and escapes them
escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g'
}

# Build locally using circle ci
circleci build \
    --config="$DIR/config.yml" \
    -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    -e AWS_DEPLOY_SECRETS_ID="$AWS_DEPLOY_SECRETS_ID" \
    -e AWS_DEPLOY_SECRETS_KEY="$AWS_DEPLOY_SECRETS_KEY" \
    -e machine_user_key="$(escape_newlines "$(cat $DIR/../machine_user_key)")" \
    -e SC_NODE_EC2_PEM="$(cat $HOME/.ssh/sc-node-ec2.pem | base64)"
