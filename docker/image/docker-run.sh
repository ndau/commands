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
    port="$1"
    data_dir="$2"
    echo "Running redis..."

    redis-server --dir "$data_dir" \
                 --port "$port" \
                 --save 60 1 \
                 >"$LOG_DIR/redis.log" 2>&1 &
    wait_port "$port"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$port" ping) == "PONG" ]]
    do
        :
    done
}

run_noms() {
    port="$1"
    data_dir="$2"
    echo "Running noms..."

    ./noms serve --port="$port" "$data_dir" \
           >"$LOG_DIR/noms.log" 2>&1 &
    wait_port "$port"
}

run_node() {
    port="$1"
    redis_port="$2"
    noms_port="$3"
    echo "Running ndaunode..."

    ./ndaunode -spec http://localhost:"$noms_port" \
                   -index localhost:"$redis_port" \
                   -addr 0.0.0.0:"$port" \
                   >"$LOG_DIR/ndaunode.log" 2>&1 &
    wait_port "$port"
}

run_tm() {
    p2p_port="$1"
    rpc_port="$2"
    node_port="$3"
    data_dir="$4"
    echo "Running tendermint..."

    CHAIN=ndau \
    ./tendermint node --home "$data_dir" \
                      --proxy_app tcp://localhost:"$node_port" \
                      --p2p.laddr tcp://0.0.0.0:"$p2p_port" \
                      --rpc.laddr tcp://0.0.0.0:"$rpc_port" \
                      --log_level="*:debug" \
                      >"$LOG_DIR/tendermint.log" 2>&1 &
    wait_port "$p2p_port"
    wait_port "$rpc_port"
}

run_ndauapi() {
    echo Running ndauapi...

    NDAUAPI_NDAU_RPC_URL=http://localhost:"$TM_RPC_PORT" \
    ./ndauapi >"$LOG_DIR/ndauapi.log" 2>&1 &
}

run_redis "$REDIS_PORT" "$REDIS_DATA_DIR"
run_noms "$NOMS_PORT" "$NOMS_DATA_DIR"
run_node "$NODE_PORT" "$REDIS_PORT" "$NOMS_PORT"
run_tm "$TM_P2P_PORT" "$TM_RPC_PORT" "$NODE_PORT" "$TM_DATA_DIR"
run_ndauapi

IDENTITY_FILE=node-identity.tgz
if [ ! -f "$SCRIPT_DIR/$IDENTITY_FILE" ]; then
    echo "Generating identity file..."

    cd "$DATA_DIR" || exit 1
    tar -czf "$SCRIPT_DIR/$IDENTITY_FILE" \
        tendermint/config/node_key.json \
        tendermint/config/priv_validator_key.json \
        tendermint/data/priv_validator_state.json

    echo "Done; run the following command to get it:"
    echo "  docker cp $NODE_ID:$SCRIPT_DIR/$IDENTITY_FILE $IDENTITY_FILE"
fi

echo "Node group $NODE_ID is now running"

# Wait forever to keep the container alive.
while true; do sleep 86400; done
