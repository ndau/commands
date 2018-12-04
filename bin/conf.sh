#!/bin/bash

SETUP_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
source $SETUP_DIR/env.sh

# Configure tendermint.
cd $TENDERMINT_DIR
./tendermint init --home $TENDERMINT_CHAOS_DATA_DIR
./tendermint init --home $TENDERMINT_NDAU_DATA_DIR
sed -i '' -E \
    -e 's/^(create_empty_blocks = .*)/#\1/' \
    -e 's/^(create_empty_blocks_interval =) (.*)/\1 300/' \
    $TENDERMINT_CHAOS_DATA_DIR/config/config.toml \
    $TENDERMINT_NDAU_DATA_DIR/config/config.toml

# Configure chaos.
$CHAOS_DIR/chaos conf

# Configure ndau.
$NDAU_DIR/ndau conf http://localhost:$TM_NDAU_RPC_PORT

# Generate and copy genesis files if they're not there already.
cd $NODE_DATA_DIR
if [ ! -e genesis.*.toml ]; then
    $CHAOS_GENESIS_DIR/generate --out .

    # reset.sh makes sure we start fresh (zero genesis files).
    # conf.sh makes sure we only create one if one's not there (wind up with one genesis file).
    # However, we are careful to get the latest one if there are somehow multiple present.
    GENESIS_TOML=`ls -t genesis.*.toml | head -n 1`
    $CHAOS_GENESIS_DIR/genesis -g $NODE_DATA_DIR/$GENESIS_TOML -n $NOMS_CHAOS_DATA_DIR

    # This is needed for things like RFE transactions to function.
    ASSC_TOML=`ls -t assc.*.toml | head -n 1`
    $NDAU_DIR/ndau conf update-from $NODE_DATA_DIR/$ASSC_TOML

    # Use this as a flag for run.sh to know whether to update ndau conf and chain with the
    # generated files.
    touch $NEEDS_UPDATE_FLAG_FILE
fi
