export NODE_CHAOS_PORT=26650
export NODE_NDAU_PORT=26651
export NOMS_CHAOS_PORT=8000
export NOMS_NDAU_PORT=8001
export REDIS_CHAOS_PORT=6379
export REDIS_NDAU_PORT=6380
export TM_CHAOS_P2P_PORT=26660
export TM_CHAOS_RPC_PORT=26670
export TM_NDAU_P2P_PORT=26661
export TM_NDAU_RPC_PORT=26671

export BIN_DIR=/image/bin
export DATA_DIR=/image/data

export NODE_DATA_DIR="$DATA_DIR"/ndauhome
export NOMS_CHAOS_DATA_DIR="$DATA_DIR"/chaos/noms
export NOMS_NDAU_DATA_DIR="$DATA_DIR"/ndau/noms
export REDIS_CHAOS_DATA_DIR="$DATA_DIR"/chaos/redis
export REDIS_NDAU_DATA_DIR="$DATA_DIR"/ndau/redis
export TM_CHAOS_DATA_DIR="$DATA_DIR"/chaos/tendermint
export TM_NDAU_DATA_DIR="$DATA_DIR"/ndau/tendermint

export NDAUHOME="$NODE_DATA_DIR"
