#!/bin/bash

pid_found=false

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

cd "$(dirname "$0")" || exit 1

if [ -z "$1" ]; then
    cmds=(ndau_tm ndau_node ndau_noms ndau_redis chaos_tm chaos_node chaos_noms chaos_redis)
else
    cmds=("$@")
fi

for cmd in "${cmds[@]}"; do
    try_kill "$cmd"
done

# Give them all a chance to shutdown before we force-kill anything.
if [ "$pid_found" = true ]; then
    sleep 1
fi

for cmd in "${cmds[@]}"; do
    force_kill "$cmd"
done

for cmd in "${cmds[@]}"; do
    check_killed "$cmd"
done
