#!/bin/bash
# This helper script returns the values used for circle ci

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# escape newlines removes newline characters and escapes them
escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g'
}

# Not included are the following AWS keys, which should be from a specified ci/cd user.
# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
# AWS_DEPLOY_SCRETS_ID
# AWS_DEPLOY_SCRETS_KEY
# HONEYCOMB_DATASET
# HONEYCOMB_KEY
# SLACK_DEPLOYS_KEY
echo machine_user_key="$(escape_newlines "$(cat $DIR/../machine_user_key)")"
echo SC_NODE_EC2_PEM="$(cat $HOME/.ssh/sc-node-ec2.pem | base64)"
