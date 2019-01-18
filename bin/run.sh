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
    node_num="$1"
    echo running redis for "chaos-$node_num"

    data_dir="$REDIS_CHAOS_DATA_DIR-$node_num"
    redis_port=$(expr "$REDIS_PORT" + 2 \* "$node_num")
    output_name="$CMDBIN_DIR/chaos_redis-$node_num"

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

#---------- noms for chaos -------------
chaos_noms() {
    node_num="$1"
    echo running noms for "chaos-$node_num"

    data_dir="$NOMS_CHAOS_DATA_DIR-$node_num"
    noms_port=$(expr "$NOMS_PORT" + 2 \* "$node_num")
    output_name="$CMDBIN_DIR/chaos_noms-$node_num"

    cd "$NOMS_DIR" || exit 1

    mkdir -p "$data_dir"
    ./noms serve --port="$noms_port" "$data_dir" >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$noms_port"
}

#---------- run chaosnode ----------
chaos_node() {
    node_num="$1"
    echo running node for "chaos-$node_num"

    ndau_home="$NODE_DATA_DIR-$node_num"
    port_offset=$(expr 2 \* "$node_num")
    noms_port=$(expr "$NOMS_PORT" + "$port_offset")
    redis_port=$(expr "$REDIS_PORT" + "$port_offset")
    node_port=$(expr "$NODE_PORT" + "$port_offset")
    genesis_config="$TENDERMINT_CHAOS_DATA_DIR-$node_num/config/genesis"
    output_name="$CMDBIN_DIR/chaos_node-$node_num"

    cd "$COMMANDS_DIR" || exit 1

    #---------- get app hash from chaosnode ----------
    echo "  getting chaosnode app hash"
    chaos_hash=$(NDAUHOME="$ndau_home" \
        ./chaosnode -spec http://localhost:"$noms_port" -echo-hash 2>/dev/null)
    # jq doesn't support an inplace operation
    jq ".app_hash= if .app_hash == \"\" then \"$chaos_hash\" else .app_hash end" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"

    echo "  launching chaosnode"
    HONEYCOMB_DATASET=chaos-dev \
    NDAUHOME="$ndau_home" \
    ./chaosnode -spec http://localhost:"$noms_port" \
                -index localhost:"$redis_port" \
                -addr 0.0.0.0:"$node_port" \
                >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$node_port"
}

#---------- run chaos tendermint ----------
chaos_tm() {
    node_num="$1"
    echo running tendermint for "chaos-$node_num"

    data_dir="$TENDERMINT_CHAOS_DATA_DIR-$node_num"
    port_offset=$(expr 2 \* "$node_num")
    node_port=$(expr "$NODE_PORT" + "$port_offset")
    p2p_port=$(expr "$TM_P2P_PORT" + "$port_offset")
    rpc_port=$(expr "$TM_RPC_PORT" + "$port_offset")
    output_name="$CMDBIN_DIR/chaos_tm-$node_num"

    cd "$TENDERMINT_DIR" || exit 1

    HONEYCOMB_DATASET=chaos-tm-dev \
    ./tendermint node --home "$data_dir" \
                      --proxy_app tcp://localhost:"$node_port" \
                      --p2p.laddr tcp://0.0.0.0:"$p2p_port" \
                      --rpc.laddr tcp://0.0.0.0:"$rpc_port" \
                      >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$rpc_port"
    wait_port "$p2p_port"

    echo "  ./chaos conf \"http://localhost:$rpc_port\""
}

#---------- redis for ndau -------------
ndau_redis() {
    node_num="$1"
    echo running redis for "ndau-$node_num"

    data_dir="$REDIS_NDAU_DATA_DIR-$node_num"
    redis_port=$(expr "$REDIS_PORT" + 2 \* "$node_num" + 1)
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
    noms_port=$(expr "$NOMS_PORT" + 2 \* "$node_num" + 1)
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
    port_offset=$(expr 2 \* "$node_num" + 1)
    noms_port=$(expr "$NOMS_PORT" + "$port_offset")
    redis_port=$(expr "$REDIS_PORT" + "$port_offset")
    node_port=$(expr "$NODE_PORT" + "$port_offset")
    chaos_rpc_port=$(expr "$TM_RPC_PORT" + "$port_offset" - 1)
    genesis_config="$TENDERMINT_NDAU_DATA_DIR-$node_num/config/genesis"
    output_name="$CMDBIN_DIR/ndau_node-$node_num"

    cd "$COMMANDS_DIR" || exit 1

    # Import genesis data if we haven't already.
    if [ -e "$NEEDS_UPDATE_FLAG_FILE-$node_num" ]; then
        echo "  updating ndau config using $GENESIS_TOML"
        NDAUHOME="$ndau_home" \
        ./ndaunode -spec http://localhost:"$noms_port" \
                   -index localhost:"$redis_port" \
                   -update-conf-from "$GENESIS_TOML"

        # The config toml file has now been generated, edit it.
        sed -i '' \
            -e "s@ChaosAddress = \".*\"@ChaosAddress = \"http://localhost:$chaos_rpc_port\"@" \
            "$ndau_home/ndau/config.toml"

        echo "  updating ndau chain using $ASSC_TOML"
        NDAUHOME="$ndau_home" \
        ./ndaunode -spec http://localhost:"$noms_port" \
                   -index localhost:"$redis_port" \
                   -update-chain-from "$ASSC_TOML"

        # We've updated, remove the flag file so we don't update again on the next run.
        rm "$NEEDS_UPDATE_FLAG_FILE-$node_num"
    fi

    #---------- get app hash from ndaunode ----------
    echo "  getting ndaunode app hash"
    ndau_hash=$(NDAUHOME="$ndau_home" \
        ./ndaunode -spec http://localhost:"$noms_port" -echo-hash 2>/dev/null)
    # jq doesn't support an inplace operation
    jq ".app_hash= if .app_hash == \"\" then \"$ndau_hash\" else .app_hash end" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"

    echo "  launching ndaunode"
    HONEYCOMB_DATASET=ndau-dev \
    NDAUHOME="$ndau_home" \
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
    port_offset=$(expr 2 \* "$node_num" + 1)
    node_port=$(expr "$NODE_PORT" + "$port_offset")
    p2p_port=$(expr "$TM_P2P_PORT" + "$port_offset")
    rpc_port=$(expr "$TM_RPC_PORT" + "$port_offset")
    output_name="$CMDBIN_DIR/ndau_tm-$node_num"

    cd "$TENDERMINT_DIR" || exit 1

    HONEYCOMB_DATASET=ndau-tm-dev \
    ./tendermint node --home "$data_dir" \
                      --proxy_app tcp://localhost:"$node_port" \
                      --p2p.laddr tcp://0.0.0.0:"$p2p_port" \
                      --rpc.laddr tcp://0.0.0.0:"$rpc_port" \
                      >"$output_name.log" 2>&1 &
    echo $! >"$output_name.pid"
    wait_port "$rpc_port"
    wait_port "$p2p_port"

    echo "  ./ndau conf \"http://localhost:$rpc_port\""
}

finalize() {
    cd "$COMMANDS_DIR" || exit 1

    if [ -e "$NEEDS_UPDATE_FLAG_FILE" ]; then
        # We only update the 0'th node's config.  This is because the account claim step below
        # affects the blockchain.  It gets propagated to the other nodes' blockchains, but their
        # ndau and chaos tool configs don't get updated.  This is okay, since developers always
        # use the ndau-0 directory as NDAUHOME when running chaos and ndau tool commands.  The
        # other nodes' config files will simply sit there, dormant.  We could even make it so
        # they are not there at all, but they were needed earlier by ndau_node() for each node,
        # so we leave them there.  They are valid, but not useable for getting/setting sysvars.
        ndau_home="$NODE_DATA_DIR-0"

        # Claim the bpc operations account.  This puts the validation keys into ndautool.toml.
        NDAUHOME="$ndau_home" ./ndau account claim "$BPC_OPS_ACCT_NAME"

        # Copy the bpc keys to the chaos tool toml file under the sysvar identity.
        NDAUHOME="$ndau_home" ./chaos id copy-keys-from "$SYSVAR_ID" "$BPC_OPS_ACCT_NAME" 

        # We've updated, remove the flag file so we don't update again on the next run.
        rm "$NEEDS_UPDATE_FLAG_FILE"
    fi
}

if [ -z "$1" ]; then
    initialize

    # Kill everything first.  It's too easy to forget the ./kill.sh between test runs.
    "$CMDBIN_DIR"/kill.sh

    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        chaos_redis "$node_num"
        chaos_noms "$node_num"
        chaos_node "$node_num"
        chaos_tm "$node_num"
        ndau_redis "$node_num"
        ndau_noms "$node_num"
        ndau_node "$node_num"
        ndau_tm "$node_num"
    done

    finalize
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
    finalize
fi

echo "done."
