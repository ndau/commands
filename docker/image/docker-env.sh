#!/bin/bash

export NODE_PORT=26650
export NOMS_PORT=8000
export REDIS_PORT=6379
export TM_P2P_PORT=26660
export TM_RPC_PORT=26670
export NDAUAPI_PORT=3030

export ULIMIT_AMOUNT=1024
export REDIS_MAX_CLIENTS=$((ULIMIT_AMOUNT - 32))

export ROOT_DIR=/image
export BIN_DIR="$ROOT_DIR"/bin
export DATA_DIR="$ROOT_DIR"/data
export LOG_DIR="$ROOT_DIR"/logs

export NODE_DATA_DIR="$DATA_DIR"/ndau
export NOMS_DATA_DIR="$DATA_DIR"/noms
export REDIS_DATA_DIR="$DATA_DIR"/redis
export TM_DATA_DIR="$DATA_DIR"/tendermint

export SYSTEM_VARS_TOML="$ROOT_DIR/system_vars.toml"
export SYSTEM_ACCOUNTS_TOML="$ROOT_DIR/system_accounts.toml"

export NDAUHOME="$NODE_DATA_DIR"

export SNAPSHOT_URL="https://s3.amazonaws.com"
export SNAPSHOT_BUCKET="ndau-snapshots"
export GENERATED_GENESIS_SNAPSHOT="*"
export LOCAL_SNAPSHOT="$ROOT_DIR/snapshot-$NETWORK-0.tgz"

