#!/bin/bash
# This helper script returns the values used for circle ci

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# escape newlines removes newline characters and escapes them
escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e's/\n/\\n/g'
}

# Not included are the following AWS keys, which should be from a specified ci/cd user.
# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
echo gomu="$(escape_newlines "$(cat $DIR/../gomu)")"
echo helm_ca_pem="$(escape_newlines "$(cat ~/.helm/ca.pem)")"
echo helm_cert_pem="$(escape_newlines "$(cat ~/.helm/cert.pem)")"
echo helm_key_pem="$(escape_newlines "$(cat ~/.helm/key.pem)")"
echo kube_config="$(escape_newlines "$(cat ~/.kube/dev-chaos.yaml)")"
