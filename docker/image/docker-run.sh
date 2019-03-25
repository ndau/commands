SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

echo "Running $NODE_ID node group..."

# If there's no data directory yet, it means we're starting from scratch.
if [ ! -d "$DATA_DIR" ]; then
    echo "Configuring node group..."
    /bin/bash "$SCRIPT_DIR"/docker-conf.sh
fi

# This is needed because in the long term, noms eats more than 256 file descriptors
ulimit -n 1024

# All commands are run out of the bin directory.
cd "$BIN_DIR" || exit 1

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"

wait_port() {
    # Block until the given port becomes open.
    until nc -z localhost "$1" 2>/dev/null
    do
        :
    done
}

run_redis() {
    chain="$1"
    port="$2"
    data_dir="$3"
    echo "Running $chain redis..."

    redis-server --dir "$data_dir" \
                 --port "$port" \
                 --save 60 1 \
                 >"$LOG_DIR/${chain}_redis.log" 2>&1 &
    wait_port "$port"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$port" ping) == "PONG" ]]
    do
        :
    done
}

run_noms() {
    chain="$1"
    port="$2"
    data_dir="$3"
    echo "Running $chain noms..."

    ./noms serve --port="$port" "$data_dir" \
           >"$LOG_DIR/${chain}_noms.log" 2>&1 &
    wait_port "$port"
}

run_node() {
    chain="$1"
    port="$2"
    redis_port="$3"
    noms_port="$4"
    echo "Running $chain node..."

    chainnode="${chain}node"
    ./"$chainnode" -spec http://localhost:"$noms_port" \
                   -index localhost:"$redis_port" \
                   -addr 0.0.0.0:"$port" \
                   >"$LOG_DIR/${chain}_node.log" 2>&1 &
    wait_port "$port"
}

run_tm() {
    chain="$1"
    p2p_port="$2"
    rpc_port="$3"
    node_port="$4"
    data_dir="$5"
    echo "Running $chain tendermint..."

    CHAIN="$chain" \
    ./tendermint node --home "$data_dir" \
                      --proxy_app tcp://localhost:"$node_port" \
                      --p2p.laddr tcp://0.0.0.0:"$p2p_port" \
                      --rpc.laddr tcp://0.0.0.0:"$rpc_port" \
                      --log_level="*:debug" \
                      >"$LOG_DIR/${chain}_tm.log" 2>&1 &
    wait_port "$p2p_port"
    wait_port "$rpc_port"
}

run_ndauapi() {
    echo Running ndauapi...

    NDAUAPI_CHAOS_RPC_URL=http://localhost:"$TM_CHAOS_RPC_PORT" \
    NDAUAPI_NDAU_RPC_URL=http://localhost:"$TM_NDAU_RPC_PORT" \
    ./ndauapi >"$LOG_DIR/ndauapi.log" 2>&1 &
}

run_redis chaos "$REDIS_CHAOS_PORT" "$REDIS_CHAOS_DATA_DIR"
run_noms chaos "$NOMS_CHAOS_PORT" "$NOMS_CHAOS_DATA_DIR"
run_node chaos "$NODE_CHAOS_PORT" "$REDIS_CHAOS_PORT" "$NOMS_CHAOS_PORT"
run_tm chaos "$TM_CHAOS_P2P_PORT" "$TM_CHAOS_RPC_PORT" "$NODE_CHAOS_PORT" "$TM_CHAOS_DATA_DIR"

run_redis ndau "$REDIS_NDAU_PORT" "$REDIS_NDAU_DATA_DIR"
run_noms ndau "$NOMS_NDAU_PORT" "$NOMS_NDAU_DATA_DIR"
run_node ndau "$NODE_NDAU_PORT" "$REDIS_NDAU_PORT" "$NOMS_NDAU_PORT"
run_tm ndau "$TM_NDAU_P2P_PORT" "$TM_NDAU_RPC_PORT" "$NODE_NDAU_PORT" "$TM_NDAU_DATA_DIR"

run_ndauapi

echo "Node group $NODE_ID is now running"

# Wait forever to keep the container alive.
while true; do sleep 86400; done
