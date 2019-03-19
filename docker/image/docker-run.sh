SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

REDIS_CHAOS_PORT=6379
REDIS_NDAU_PORT=6380
REDIS_CHAOS_DATA_DIR="$DATA_DIR"/redis-chaos
REDIS_NDAU_DATA_DIR="$DATA_DIR"/redis-ndau

# This is needed because in the long term, noms eats more than 256 file descriptors
ulimit -n 1024

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
                 2>&1 &
    wait_port "$REDIS_CHAOS_PORT"

    # Redis isn't really ready when it's port is open, wait for a ping to work.
    until [[ $(redis-cli -p "$REDIS_CHAOS_PORT" ping) == "PONG" ]]
    do
        :
    done
}

chaos_redis

echo ndau node group is now running

# Wait forever to keep the container alive.
while true; do sleep 86400; done
