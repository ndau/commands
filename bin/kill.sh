#!/bin/bash

# This will get set to true if we found any pid and attempted to kill it gracefully.
pid_found=false

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    cd "$CMDBIN_DIR" || exit 1
}

# This tries to kill a task nicely and does not wait.
try_kill() {
    if [ -e "$1".pid ]; then
        pid=$(cat "$1".pid)
        if ps -p "$pid" > /dev/null; then
            echo killing "$1"
            kill "$pid"
        fi
        pid_found=true
    else
        echo "skipping $1 ($1.pid not found)"
    fi
}

# This will force-kill using -9 and wait 1 second for it.
force_kill() {
    if [ -e "$1".pid ]; then
        pid=$(cat "$1".pid)
        if ps -p "$pid" > /dev/null; then
            echo force killing "$1"
            kill -9 "$pid"
            sleep 1
        fi
    fi
}

# This logs whether or not we killed a task.
check_killed() {
    if [ -e "$1".pid ]; then
        pid=$(cat "$1".pid)
        if ps -p "$pid" > /dev/null; then
            echo "process $pid ($1) won't die"
        else
            rm "$1".pid
            echo "$1" killed
        fi
    fi
}

initialize

if [ -z "$1" ]; then
    cmds=(ndau_tm ndau_node ndau_noms ndau_redis)
    while IFS=$'\n' read -r line; do node_nums+=("$line"); done < <(seq "$HIGH_NODE_NUM" 0)
else
    # We support killing a single process for a given node.
    cmds=("$1")
    node_num="$2"

    # Default to the first node in a single-node localnet.
    if [ -z "$node_num" ]; then
        node_num=0
    fi

    node_nums=("$node_num")
fi

for node_num in "${node_nums[@]}";
do
    for cmd in "${cmds[@]}"; do
        try_kill "$cmd-$node_num"
    done
done

# Give them all a chance to shutdown before we force-kill anything.
if [ "$pid_found" = true ]; then
    sleep 1
fi

for node_num in "${node_nums[@]}";
do
    for cmd in "${cmds[@]}"; do
        force_kill "$cmd-$node_num"
    done
done

for node_num in "${node_nums[@]}";
do
    for cmd in "${cmds[@]}"; do
        check_killed "$cmd-$node_num"
    done
done
