#!/bin/bash

# checks to see if a task exists; if so, shows its status
checkstatus() {
    if [ -e ndau_tm.pid ]; then
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

cd "$(dirname "$0")" || exit 1

if [ -n "$1" ]; then
    echo "hi x $1 x"
    checkstatus "$1"
else
    checkstatus chaos_redis
    checkstatus chaos_noms
    checkstatus chaos_node
    checkstatus chaos_tm
    checkstatus ndau_redis
    checkstatus ndau_noms
    checkstatus ndau_node
    checkstatus ndau_tm
fi
