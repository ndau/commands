#!/bin/bash

export NODE_PORT=26650
export NOMS_PORT=8000
export REDIS_PORT=6379
export TM_P2P_PORT=26660
export TM_RPC_PORT=26670
export NDAUAPI_PORT=3030
export PG_PORT=5432

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
export PGDATA="$DATA_DIR"/postgres

export SYSTEM_VARS_TOML="$ROOT_DIR/system_vars.toml"
export SYSTEM_ACCOUNTS_TOML="$ROOT_DIR/system_accounts.toml"

export NDAUHOME="$NODE_DATA_DIR"

export SNAPSHOT_URL="https://s3.amazonaws.com"
export SNAPSHOT_BUCKET="ndau-snapshots"
export GENERATED_GENESIS_SNAPSHOT="*"
export LOCAL_SNAPSHOT="$ROOT_DIR/snapshot-$NETWORK-0.tgz"

export WEBHOOK_URL="https://7ovwffck3i.execute-api.us-east-1.amazonaws.com/$NETWORK/claim_winner"

# Set up the postgres master password. There are a few factors to consider for this:
#
# 1. Postgres is set up with trust authentication for socket connections, so anyone
#    who can shell into the container has root access regardless.
# 2. This should be a random value at container init, but should be stable during
#    the lifetime of the container.
#
# Together, those two factors mean that it's both safe and necessary to generate
# this password from a file.
postgres_pw_file="/image/postgres-pw"
if [ ! -s "$postgres_pw_file" ]; then
    head -c 12 /dev/urandom | base64 > "$postgres_pw_file"
    chmod u=r,g=,o= "$postgres_pw_file"
    chown postgres:postgres "$postgres_pw_file"
fi
# we can still export the variable, because this script is run as root
POSTGRES_PASSWORD=$(cat "$postgres_pw_file")
export POSTGRES_PASSWORD
