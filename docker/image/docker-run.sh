SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

NODE_CHAOS_PORT=26650
NODE_NDAU_PORT=26651
NOMS_CHAOS_PORT=8000
NOMS_NDAU_PORT=8001
REDIS_CHAOS_PORT=6379
REDIS_NDAU_PORT=6380
REDIS_CHAOS_DATA_DIR="$DATA_DIR"/redis-chaos
REDIS_NDAU_DATA_DIR="$DATA_DIR"/redis-ndau

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

chaos_redis() {
    echo Running chaos redis...

    mkdir -p "$REDIS_CHAOS_DATA_DIR"
    redis-server --dir "$REDIS_CHAOS_DATA_DIR" \
                 --port "$REDIS_CHAOS_PORT" \
                 --save 60 1 \
                 >"$LOG_DIR/chaos_redis.log" 2>&1 &
    wait_port "$REDIS_CHAOS_PORT"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$REDIS_CHAOS_PORT" ping) == "PONG" ]]
    do
        :
    done
}

chaos_noms() {
    echo Running chaos noms...

    ./noms serve --port="$NOMS_CHAOS_PORT" "$NOMS_CHAOS_DATA_DIR" \
           >"$LOG_DIR/chaos_noms.log" 2>&1 &
    wait_port "$NOMS_CHAOS_PORT"
}

chaos_node() {
    echo Running chaos node...

    echo "  getting chaosnode app hash"
    chaos_hash=$(./chaosnode -spec http://localhost:"$NOMS_CHAOS_PORT" -echo-hash 2>/dev/null)
    # Set the app hash if it's not already set.
    sed -i -E \
        -e 's/"app_hash": ""/"app_hash": "'$chaos_hash'"/' \
        "$TENDERMINT_CHAOS_DATA_DIR/config/genesis.json"

    echo "  launching chaosnode"
    ./chaosnode -spec http://localhost:"$NOMS_CHAOS_PORT" \
                -index localhost:"$REDIS_CHAOS_PORT" \
                -addr 0.0.0.0:"$NODE_CHAOS_PORT" \
                >"$LOG_DIR/chaos_node.log" 2>&1 &
    wait_port "$NODE_CHAOS_PORT"
}

chaos_tm() {
    echo Running chaos tendermint...

    CHAIN=chaos \
    ./tendermint node --home "$TENDERMINT_CHAOS_DATA_DIR" \
                      --proxy_app tcp://localhost:"$NODE_CHAOS_PORT" \
                      --p2p.laddr tcp://0.0.0.0:"$TM_CHAOS_P2P_PORT" \
                      --rpc.laddr tcp://0.0.0.0:"$TM_CHAOS_RPC_PORT" \
                      --log_level="*:debug" \
                      >"$LOG_DIR/chaos_tm.log" 2>&1 &
    wait_port "$TM_CHAOS_RPC_PORT"
    wait_port "$TM_CHAOS_P2P_PORT"
}

chaos_redis
chaos_noms
chaos_node
chaos_tm

echo "$NODE_ID" is now running

# Wait forever to keep the container alive.
while true; do sleep 86400; done
