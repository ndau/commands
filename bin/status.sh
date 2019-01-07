#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    cd "$CMDBIN_DIR" || exit 1
}

# checks to see if a task exists; if so, shows its status
checkstatus() {
    if [ -e "$1".pid ]; then
        pid=$(cat "$1".pid)
        if ps -p "$pid" > /dev/null; then
            ps -p "$pid" | tail -1
        else
            echo "$1.pid exists but task $pid is not running"
        fi
    else
        echo "$1.pid does not exist; $1 is probably not running"
    fi
}

initialize

if [ -n "$1" ]; then
    # We support checking a single process for a given node.
    cmd="$1"
    node_num="$2"

    # Default to the first node in a single-node localnet.
    if [ -z "$node_num" ]; then
        node_num=0
    fi

    checkstatus "$cmd-$node_num"
else
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        checkstatus "chaos_redis-$node_num"
        checkstatus "chaos_noms-$node_num"
        checkstatus "chaos_node-$node_num"
        checkstatus "chaos_tm-$node_num"
        checkstatus "ndau_redis-$node_num"
        checkstatus "ndau_noms-$node_num"
        checkstatus "ndau_node-$node_num"
        checkstatus "ndau_tm-$node_num"
    done
fi
