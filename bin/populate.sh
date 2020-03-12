#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/ndau/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    # We use the home directory of the 0'th node, even in a multi-node localnet.
    ndau_home="$NODE_DATA_DIR-0"
    ND="$COMMANDS_DIR"/ndau
}

createaccount() {
    name=$1
    NDAUHOME="$ndau_home" "$ND" account new "$name"
}

set-validationaccount() {
    name=$1
    NDAUHOME="$ndau_home" "$ND" account set-validation "$name"
}

rfeTo() {
    name=$1
    amt=$2
    NDAUHOME="$ndau_home" "$ND" -v rfe "$amt" "$name"
}

issueNdau() {
    amt=$1
    NDAUHOME="$ndau_home" "$ND" -v issue "$amt"
}

transfer() {
    from=$1
    to=$2
    amt=$3
    NDAUHOME="$ndau_home" "$ND" transfer "$amt" "$from" "$to"
}

query() {
    name=$1
    NDAUHOME="$ndau_home" "$ND" account query "$name"
}

create() {
    createaccount alice
    createaccount bob
    createaccount carol
    createaccount drew
}

createbig() {
    createaccount xavier
    createaccount yannis
    createaccount zelda
}

issue() {
    rfeTo alice 10
    rfeTo bob 20
    rfeTo carol 30
    rfeTo drew 40
    issueNdau 100
}

issuebig() {
    rfeTo xavier 10000
    issueNdau 10000
    rfeTo yannis 1000
    issueNdau 1000
    rfeTo zelda 20000
    issueNdau 20000
}

set-validation() {
    set-validationaccount alice
    set-validationaccount bob
    set-validationaccount carol
    set-validationaccount drew
}

xfer() {
    transfer carol alice 3.14
    transfer bob alice 2.71828
    transfer drew alice 1.414
}

refund() {
    transfer alice carol 3.14
    transfer alice bob 2.71828
    transfer alice drew 1.414
}

status() {
    echo alice && query alice
    echo bob && query bob
    echo carol && query carol
    echo drew && query drew
}

if [ -z "$1" ]; then
    echo "enter a command"
else
    initialize
    "$@"
fi
