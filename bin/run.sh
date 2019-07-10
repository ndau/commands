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

#---------- redis for ndau -------------
ndau_redis() {
    node_num="$1"
    echo running redis for "ndau-$node_num"

    data_dir="$REDIS_NDAU_DATA_DIR-$node_num"
    redis_port=$((REDIS_PORT + node_num))
    output_name="$CMDBIN_DIR/ndau_redis-$node_num"

    mkdir -p "$data_dir"
    redis-server --dir "$data_dir" \
                 --port "$redis_port" \
                 --save 60 1 \
                 >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$redis_port"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$redis_port" ping) == "PONG" ]]
    do
        :
    done
}

#---------- noms for ndau -------------
ndau_noms() {
    node_num="$1"
    echo running noms for "ndau-$node_num"

    data_dir="$NOMS_NDAU_DATA_DIR-$node_num"
    noms_port=$((NOMS_PORT + node_num))
    output_name="$CMDBIN_DIR/ndau_noms-$node_num"

    cd "$NOMS_DIR" || exit 1

    mkdir -p "$data_dir"
    ./noms serve --port="$noms_port" "$data_dir" >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$noms_port"
}

#---------- run ndaunode -------------
ndau_node() {
    node_num="$1"
    echo running node for "ndau-$node_num"

    ndau_home="$NODE_DATA_DIR-$node_num"
    noms_port=$((NOMS_PORT + node_num))
    redis_port=$((REDIS_PORT + node_num))
    node_port=$((NODE_PORT + node_num))
    genesis_config="$TENDERMINT_NDAU_DATA_DIR-$node_num/config/genesis"
    output_name="$CMDBIN_DIR/ndau_node-$node_num"

    cd "$COMMANDS_DIR" || exit 1

    #---------- get app hash from ndaunode ----------
    echo "  getting ndaunode app hash"
    ndau_hash=$(NDAUHOME="$ndau_home" \
        ./ndaunode -spec http://localhost:"$noms_port" -echo-hash 2>/dev/null)
    # jq doesn't support an inplace operation
    jq ".app_hash= if .app_hash == \"\" then \"$ndau_hash\" else .app_hash end" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"

    echo "  launching ndaunode"
    NDAUHOME="$ndau_home" \
    NODE_ID="$MONIKER_PREFIX-$node_num" \
    ./ndaunode -spec http://localhost:"$noms_port" \
               -index localhost:"$redis_port" \
               -addr 0.0.0.0:"$node_port" \
               >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$node_port"
}

#---------- run ndau tendermint ----------
ndau_tm() {
    node_num="$1"
    echo running tendermint for "ndau-$node_num"

    data_dir="$TENDERMINT_NDAU_DATA_DIR-$node_num"
    node_port=$((NODE_PORT + node_num))
    p2p_port=$((TM_P2P_PORT + node_num))
    rpc_port=$((TM_RPC_PORT + node_num))
    output_name="$CMDBIN_DIR/ndau_tm-$node_num"

    cd "$TENDERMINT_DIR" || exit 1

    # https://blog.cosmos.network/one-of-the-exciting-new-features-in-0-10-0-release-is-smart-log-level-flag-e2506b4ab756
    # for details on how to configure `log_level` config variable.
    # If you're trying to debug Tendermint or asked to provide logs with debug
    # logging level, you can do so by running tendermint with
    # `--log_level="*:debug"` but you can configure individual modules differently,
    # like `--log_level="state:info,mempool:error,*:error"`.
    # value choices are debug/info/error/none
    # module options include consensus, state, p2p, mempool, proxy, node, main
    NODE_ID="$MONIKER_PREFIX-$node_num" \
    ./tendermint node --home "$data_dir" \
                      --proxy_app tcp://localhost:"$node_port" \
                      --p2p.laddr tcp://0.0.0.0:"$p2p_port" \
                      --rpc.laddr tcp://0.0.0.0:"$rpc_port" \
                      --log_format="json" \
                      --log_level="*:debug" \
                      >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    echo "  tm coming up; waiting for ports $rpc_port and $p2p_port"
    wait_port "$rpc_port"
    wait_port "$p2p_port"

    echo "  ./ndau conf \"http://localhost:$rpc_port\""
}


#---------- run ndauapi ----------
ndau_api() {
    node_num="$1"
    echo running ndauapi for "ndau-$node_num"

    api_port=$((NDAUAPI_PORT + node_num))
    output_name="$CMDBIN_DIR/ndauapi-$node_num"

    cd "$COMMANDS_DIR" || exit 1

    NDAUAPI_NDAU_RPC_URL="http://localhost:$TM_RPC_PORT" \
    NDAUAPI_PORT="$api_port" \
    ./ndauapi >"$output_name.log" \
    2>&1 &

    echo $! >"$output_name.pid"
    echo "  ndauapi coming up; waiting for port $api_port"
    wait_port "$api_port"
}

if [ -z "$1" ]; then
    initialize

    # Kill everything first.  It's too easy to forget the ./kill.sh between test runs.
    "$CMDBIN_DIR"/kill.sh

    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        ndau_redis "$node_num"
        ndau_noms "$node_num"
        ndau_node "$node_num"
        ndau_tm "$node_num"
        ndau_api "$node_num"
    done
else
    # We support running a single process for a given node.
    cmd="$1"
    node_num="$2"

    # Default to the first node in a single-node localnet.
    if [ -z "$node_num" ]; then
        node_num=0
    fi

    initialize
    "$cmd" "$node_num"
fi

echo "done."
