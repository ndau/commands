#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    ND=$NDAU_DIR/ndau
}

createaccount() {
    name=$1
    "$ND" account new "$name"
}

claimaccount() {
    name=$1
    "$ND" account claim "$name"
}

rfeTo() {
    name=$1
    amt=$2
    "$ND" -v rfe "$amt" "$name"
}

transfer() {
    from=$1
    to=$2
    amt=$3
    "$ND" transfer "$amt" "$from" "$to"
}

query() {
    name=$1
    "$ND" account query "$name"
}

create() {
    createaccount alice
    createaccount bob
    createaccount carol
    createaccount drew
}

issue() {
    rfeTo alice 10
    rfeTo bob 20
    rfeTo carol 30
    rfeTo drew 40
}

claim() {
    claimaccount alice
    claimaccount bob
    claimaccount carol
    claimaccount drew
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
