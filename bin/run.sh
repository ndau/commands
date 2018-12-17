#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    # This is needed because in the long term, noms eats more than 256 file descriptors
    ulimit -n 1024
}

wait_port() {
    # Block until the given port becomes open.
    until nc -z localhost "$1" 2>/dev/null
    do
        :
    done
}

#---------- redis for chaos -------------
chaos_redis() {
    echo running redis for chaos
    mkdir -p "$REDIS_CHAOS_DATA_DIR"
    redis-server --dir "$REDIS_CHAOS_DATA_DIR" \
                 --port "$REDIS_CHAOS_PORT" \
                 --save 60 1 \
                 >"$CMDBIN_DIR"/chaos_redis.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/chaos_redis.pid
    wait_port "$REDIS_CHAOS_PORT"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$REDIS_CHAOS_PORT" ping) == "PONG" ]]
    do
        :
    done
}

#---------- noms for chaos -------------
chaos_noms() {
    echo running noms for chaos
    cd "$NOMS_DIR" || exit 1
    mkdir -p "$NOMS_CHAOS_DATA_DIR"
    ./noms serve --port="$NOMS_CHAOS_PORT" "$NOMS_CHAOS_DATA_DIR" >"$CMDBIN_DIR"/chaos_noms.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/chaos_noms.pid
    wait_port "$NOMS_CHAOS_PORT"
}

#---------- run chaosnode ----------
chaos_node() {
    cd "$COMMANDS_DIR" || exit 1

    #---------- get app hash from chaosnode ----------
    echo getting chaosnode app hash
    CHAOS_HASH=$(./chaosnode -spec http://localhost:"$NOMS_CHAOS_PORT" -echo-hash 2>/dev/null)
    # jq doesn't support an inplace operation
    jq ".app_hash= if .app_hash == \"\" then \"$CHAOS_HASH\" else .app_hash end" \
        "$TENDERMINT_CHAOS_DATA_DIR"/config/genesis.json \
        > "$TENDERMINT_CHAOS_DATA_DIR"/config/genesis.new.json &&
        mv "$TENDERMINT_CHAOS_DATA_DIR"/config/genesis.new.json "$TENDERMINT_CHAOS_DATA_DIR"/config/genesis.json

    echo running chaosnode
    HONEYCOMB_DATASET=chaos-dev \
    ./chaosnode -spec http://localhost:"$NOMS_CHAOS_PORT" \
                -index localhost:"$REDIS_CHAOS_PORT" \
                >"$CMDBIN_DIR"/chaos_node.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/chaos_node.pid
    wait_port "$NODE_CHAOS_PORT"
}

#---------- run chaos tendermint ----------
chaos_tm() {
    echo running chaos tendermint
    cd "$TENDERMINT_DIR" || exit 1
    HONEYCOMB_DATASET=chaos-tm-dev \
    ./tendermint node --home "$TENDERMINT_CHAOS_DATA_DIR" \
                      --proxy_app tcp://localhost:"$NODE_CHAOS_PORT" \
                      --p2p.laddr tcp://0.0.0.0:"$TM_CHAOS_P2P_PORT" \
                      --rpc.laddr tcp://0.0.0.0:"$TM_CHAOS_RPC_PORT" \
                      >"$CMDBIN_DIR"/chaos_tm.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/chaos_tm.pid
    wait_port "$TM_CHAOS_RPC_PORT"
    wait_port "$TM_CHAOS_P2P_PORT"
}

#---------- redis for ndau -------------
ndau_redis() {
    echo running redis for ndau
    mkdir -p "$REDIS_NDAU_DATA_DIR"
    redis-server --dir "$REDIS_NDAU_DATA_DIR" \
                 --port "$REDIS_NDAU_PORT" \
                 --save 60 1 \
                 >"$CMDBIN_DIR"/ndau_redis.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/ndau_redis.pid
    wait_port "$REDIS_NDAU_PORT"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$REDIS_NDAU_PORT" ping) == "PONG" ]]
    do
        :
    done
}

ndau_noms() {
    #---------- noms for ndau -------------
    echo running noms for ndau
    cd "$NOMS_DIR" || exit 1
    mkdir -p "$NOMS_NDAU_DATA_DIR"
    ./noms serve --port="$NOMS_NDAU_PORT" "$NOMS_NDAU_DATA_DIR" >"$CMDBIN_DIR"/ndau_noms.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/ndau_noms.pid
    wait_port "$NOMS_NDAU_PORT"
}

ndau_node() {
    #---------- run ndaunode -------------
    cd "$COMMANDS_DIR" || exit 1

    # Import genesis data if we haven't already.
    if [ -e "$NEEDS_UPDATE_FLAG_FILE" ]; then
        # We should only have one of each of these files, but these commands get the latest ones.
        # shellcheck disable=SC2012
        GENESIS_TOML=$(ls -t "$NODE_DATA_DIR"/genesis.*.toml | head -n 1)
        # shellcheck disable=SC2012
        ASSC_TOML=$(ls -t "$NODE_DATA_DIR"/assc.*.toml | head -n 1)

        echo updating ndau config using "$GENESIS_TOML"
        ./ndaunode -spec http://localhost:"$NOMS_NDAU_PORT" \
                   -index localhost:"$REDIS_NDAU_PORT" \
                   -update-conf-from "$GENESIS_TOML"

        # The config toml file has now been generated, edit it.
        sed -i '' \
        -e "s@ChaosAddress = \".*\"@ChaosAddress = \"http://localhost:$TM_CHAOS_RPC_PORT\"@" \
        "$NODE_DATA_DIR"/ndau/config.toml

        echo updating ndau chain using "$ASSC_TOML"
        ./ndaunode -spec http://localhost:"$NOMS_NDAU_PORT" \
                   -index localhost:"$REDIS_NDAU_PORT" \
                   -update-chain-from "$ASSC_TOML"

        # We've updated, remove the flag file so we don't update again on the next run.
        rm "$NEEDS_UPDATE_FLAG_FILE"
    fi

    #---------- get app hash from ndaunode ----------
    echo getting ndaunode app hash
    NDAU_HASH=$(./ndaunode -spec http://localhost:"$NOMS_NDAU_PORT" -echo-hash 2>/dev/null)
    # jq doesn't support an inplace operation
    jq ".app_hash= if .app_hash == \"\" then \"$NDAU_HASH\" else .app_hash end" \
        "$TENDERMINT_NDAU_DATA_DIR"/config/genesis.json \
        > "$TENDERMINT_NDAU_DATA_DIR"/config/genesis.new.json &&
        mv "$TENDERMINT_NDAU_DATA_DIR"/config/genesis.new.json "$TENDERMINT_NDAU_DATA_DIR"/config/genesis.json

    # now we can run ndaunode
    echo running ndaunode
    HONEYCOMB_DATASET=ndau-dev \
    ./ndaunode -spec http://localhost:"$NOMS_NDAU_PORT" \
               -index localhost:"$REDIS_NDAU_PORT" \
               -addr 0.0.0.0:"$NODE_NDAU_PORT" \
               >"$CMDBIN_DIR"/ndau_node.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/ndau_node.pid
    wait_port "$NODE_NDAU_PORT"
}

ndau_tm() {
    #---------- run ndau tendermint ----------
    echo running ndau tendermint

    cd "$TENDERMINT_DIR" || exit 1
    HONEYCOMB_DATASET=ndau-tm-dev \
    ./tendermint node --home "$TENDERMINT_NDAU_DATA_DIR" \
                      --proxy_app tcp://localhost:"$NODE_NDAU_PORT" \
                      --p2p.laddr tcp://0.0.0.0:"$TM_NDAU_P2P_PORT" \
                      --rpc.laddr tcp://0.0.0.0:"$TM_NDAU_RPC_PORT" \
                      >"$CMDBIN_DIR"/ndau_tm.log 2>&1 &
    echo $! >"$CMDBIN_DIR"/ndau_tm.pid
    wait_port "$TM_NDAU_RPC_PORT"
    wait_port "$TM_NDAU_P2P_PORT"
}

if [ -z "$1" ]; then
    initialize

    # Kill everything first.  It's too easy to forget the ./kill.sh between test runs.
    "$CMDBIN_DIR"/kill.sh

    chaos_redis
    chaos_noms
    chaos_node
    chaos_tm
    ndau_redis
    ndau_noms
    ndau_node
    ndau_tm
else
    initialize
    for cmd in "$@"; do
        echo running "$cmd"
        "$cmd"
    done
fi

echo "done."
