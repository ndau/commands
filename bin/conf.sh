#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Configure tendermint.
cd "$TENDERMINT_DIR" || exit 1
./tendermint init --home "$TENDERMINT_CHAOS_DATA_DIR"
./tendermint init --home "$TENDERMINT_NDAU_DATA_DIR"
sed -i '' -E \
    -e 's/^(create_empty_blocks = .*)/# \1/' \
    -e 's/^(create_empty_blocks_interval =) (.*)/\1 300/' \
    "$TENDERMINT_CHAOS_DATA_DIR"/config/config.toml \
    "$TENDERMINT_NDAU_DATA_DIR"/config/config.toml

ndau_rpc_addr="http://localhost:$TM_NDAU_RPC_PORT"

# Configure chaos.
"$COMMANDS_DIR"/chaos conf
"$COMMANDS_DIR"/chaosnode --set-ndaunode "$ndau_rpc_addr"

# Configure ndau.
"$COMMANDS_DIR"/ndau conf "$ndau_rpc_addr"

# Generate and copy genesis files if they're not there already.
cd "$NODE_DATA_DIR" || exit 1

# can't test if a glob matches anything directly: https://github.com/koalaman/shellcheck/wiki/SC2144
gexists=0
for gfile in genesis.*.toml; do
    if [ -e "$gfile" ]; then
        gexists=1
        break
    fi
done

if [ "$gexists" == 0 ]; then
    "$COMMANDS_DIR"/generate --out .

    # reset.sh makes sure we start fresh (zero genesis files).
    # conf.sh makes sure we only create one if one's not there (wind up with one genesis file).
    # However, we are careful to get the latest one if there are somehow multiple present.
    #
    # Shellcheck suggests using find instead of ls, because ls sometimes transforms
    # non-ascii filenames. However, we know that won't happen with this glob, and
    # also find doesn't have an easy equivalent to `ls -t`, so we ignore that.
    #
    # shellcheck disable=SC2012
    GENESIS_TOML=$(ls -t genesis.*.toml | head -n 1)
    "$COMMANDS_DIR"/genesis -g "$NODE_DATA_DIR/$GENESIS_TOML" -n "$NOMS_CHAOS_DATA_DIR"

    # This is needed for things like RFE transactions to function.
    # shellcheck disable=SC2012
    ASSC_TOML=$(ls -t assc.*.toml | head -n 1)
    "$COMMANDS_DIR"/ndau conf update-from "$NODE_DATA_DIR/$ASSC_TOML"

    # Use this as a flag for run.sh to know whether to update ndau conf and chain with the
    # generated files.
    touch "$NEEDS_UPDATE_FLAG_FILE"
fi
