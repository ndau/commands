#!/bin/bash
# This script runs the circle ci build on your local machine.
# You can install the circle ci commandline tool with the following commands:
#

# errcho echos to stderr
errcho(){
    echo -e "$@" >&2
}

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$(which circleci)" ]; then
    errcho "Installing circleci"
    curl -o /usr/local/bin/circleci https://circle-downloads.s3.amazonaws.com/releases/build_agent_wrapper/circleci
    chmod +x /usr/local/bin/circleci
else
    errcho "circleci already installed."
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

if [ ! -f "$DIR/../machine_user_key" ]; then
    errors=true
    errcho "Missing $DIR/../machine_user_key"
fi

if [ ! -f "$HOME/.helm/ca.pem" ]; then
    errors=true
    errcho "Missing $HOME/.helm/ca.pem"
fi

if [ ! -f "$HOME/.helm/cert.pem" ]; then
    errors=true
    errcho "Missing $HOME/.helm/cert.pem"
fi

if [ ! -f "$HOME/.helm/key.pem" ]; then
    errors=true
    errcho "Missing $HOME/.helm/key.pem"
fi

if [ ! -f "$HOME/.kube/dev-chaos.yaml" ]; then
    errors=true
    errcho "Missing $HOME/.kube/dev-chaos.yaml"
fi

# exit if there's any errors
$errors && exit 1

# escape newlines removes newline characters and escapes them
escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g'
}

# Build locally using circle ci
circleci build \
    -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    -e gomu="$(escape_newlines "$(cat $DIR/../machine_user_key)")" \
    -e helm_ca_pem="$(escape_newlines "$(cat ~/.helm/ca.pem)")" \
    -e helm_cert_pem="$(escape_newlines "$(cat ~/.helm/cert.pem)")" \
    -e helm_key_pem="$(escape_newlines "$(cat ~/.helm/key.pem)")" \
    -e kube_config="$(escape_newlines "$(cat ~/.kube/dev-chaos.yaml)")"
