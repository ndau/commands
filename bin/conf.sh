#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Protection against conf.sh being run multiple times.
# We only want to flag for needs-update if we're being called from setup.sh or reset.sh.
NEEDS_UPDATE=0

# By default, this script only updates the ndau node configuration
# files in the individual ndauhomes. However, it can sometimes be useful to
# update the configuration files at the default ndauhome as well to point to
# localnet node 0, for ease of usage. This flag tracks whether we should perform
# that update.
UPDATE_DEFAULT_NDAUHOME=0

# Process command line arguments.
ARGS=("$@")
for arg in "${ARGS[@]}"; do
    if [ "$arg" = "--needs-update" ]; then
        NEEDS_UPDATE=1
    fi
    if [[ "$arg" = "--update-default-ndauhome" || "$arg" = "-U" ]]; then
        UPDATE_DEFAULT_NDAUHOME=1
    fi
done

echo Configuring tendermint...
cd "$TENDERMINT_DIR" || exit 1

for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"

    ./tendermint init --home "$tm_ndau_home"

    sed -i '' -E \
        -e 's/^(create_empty_blocks = .*)/# \1/' \
        -e 's/^(create_empty_blocks_interval =) (.*)/\1 "300s"/' \
        -e 's/^(addr_book_strict =) (.*)/\1 false/' \
        -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
        -e 's/^(moniker =) (.*)/\1 \"'"$MONIKER_PREFIX"'-'"$node_num"'\"/' \
        "$tm_ndau_home/config/config.toml"
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

# Point tendermint nodes to each other if there are more than one node in the localnet.
if [ "$NODE_COUNT" -gt 1 ]; then
    # Because of Tendermint's PeX feature, each node could gossip known peers to the others.
    # So for every node's config, we'd only need to tell it about one other node, not all of
    # them.  The last node therefore wouldn't need to know about any peers, because the
    # previous one will dial it up as a peer.  However, to be more like how things are done in
    # the automation repo, we share all peers with each other.
    ndau_peers=()
    ndau_addresses=()
    ndau_pub_keys=()

    # Build the full list of peers.
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"
        tm_ndau_priv="$tm_ndau_home/config/priv_validator_key.json"

        peer_id=$(./tendermint show_node_id --home "$tm_ndau_home")
        peer_port=$((TM_P2P_PORT + 2 * node_num + 1))
        peer="$peer_id@127.0.0.1:$peer_port"
        ndau_peers+=("$peer")

        ndau_addresses+=("$(jq -c .address "$tm_ndau_priv")")
        ndau_pub_keys+=("$(jq -c .pub_key "$tm_ndau_priv")")
    done

    # Share the peer list with every node (minus each node's own peer id).
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"
        tm_ndau_genesis="$tm_ndau_home/config/genesis.json"

        non_self_peers=("${ndau_peers[@]}")
        unset 'non_self_peers[$node_num]'
        peers=$(join_by , "${non_self_peers[@]}")
        sed -i '' -E \
            -e 's/^(persistent_peers =) (.*)/\1 \"'"$peers"'\"/' \
            "$tm_ndau_home/config/config.toml"

        # Make every node's genesis file have all nodes set up as validators.
        if [ "$node_num" = 0 ]; then
            # Construct the validator list from scratch for node 0.
            jq ".validators = []" \
               "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
                mv "$tm_ndau_genesis.new" "$tm_ndau_genesis"

            for peer_num in $(seq 0 "$HIGH_NODE_NUM");
            do
                a=${ndau_addresses[$peer_num]}
                k=${ndau_pub_keys[$peer_num]}
                p=10
                n="ndau-$peer_num"
                jq ".validators+=[{\"address\":$a,\"pub_key\":$k,\"power\":\"$p\",\"name\":\"$n\"}]" \
                   "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
                    mv "$tm_ndau_genesis.new" "$tm_ndau_genesis"
            done
        else
            # Copy the entire genesis.json files from node 0 to all other nodes.
            cp "$TENDERMINT_NDAU_DATA_DIR-0/config/genesis.json" "$tm_ndau_genesis"
        fi
    done
fi

echo Configuring ndau...
cd "$COMMANDS_DIR" || exit 1

for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    ndau_home="$NODE_DATA_DIR-$node_num"
    port_offset=$((2 * node_num))
    ndau_rpc_port=$((TM_RPC_PORT + port_offset + 1))
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    NDAUHOME="$ndau_home" ./ndau conf "$ndau_rpc_addr"
done

# Make sure the genesif files exist, since steps after this require them.
if [ ! -f "$SYSTEM_VARS_TOML" ] || [ ! -f "$SYSTEM_ACCOUNTS_TOML" ]; then
    mkdir -p "$GENESIS_FILES_DIR"
    ./generate -v -g "$SYSTEM_VARS_TOML" -a "$SYSTEM_ACCOUNTS_TOML"
fi

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" ]]; then
    node_num=0
    port_offset=$((2 * node_num))
    ndau_rpc_port=$((TM_RPC_PORT + port_offset + 1))
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    ./ndau conf "$ndau_rpc_addr"
    ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"
fi

# Use this as a flag for run.sh to know whether to update ndau conf and chain with the
# genesis files, claim bpc account, etc.
if [ "$NEEDS_UPDATE" != 0 ]; then
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        ndau_home="$NODE_DATA_DIR-$node_num"

        NDAUHOME="$ndau_home" ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"

        # For deterministic bpc account address/keys, we recover a special account with 12 eyes.
        # Since this is only for localnet/devnet/testnet (i.e. not mainnet), this is safe.
        NDAUHOME="$ndau_home" ./ndau account recover "$BPC_OPS_ACCT_NAME" \
            eye eye eye eye eye eye eye eye eye eye eye eye

        # Generate noms data for ndau node 0, copy from node 0 otherwise.
        data_dir="$NOMS_NDAU_DATA_DIR-$node_num"
        if [ "$node_num" = 0 ]; then
            NDAUHOME="$ndau_home" \
            ./ndaunode -use-ndauhome \
                       -genesisfile "$SYSTEM_VARS_TOML" \
                       -asscfile "$SYSTEM_ACCOUNTS_TOML"
            mv "$ndau_home/ndau/noms" "$data_dir"

            # set var below if ETL step is to be run. this needs to be here because ETL needs to
            # push data direct to noms dir and before noms starts
            if [ "$RUN_ETL" = "1" ]; then
                "$CMDBIN_DIR"/etl.sh $node_num
            fi
        else
            echo "  copying ndau noms from node 0 to node $node_num"
            cp -r "$NOMS_NDAU_DATA_DIR-0" "$data_dir"
        fi
    done

    # The no-node-num form of the needs-update file flags that we need to claim the bpc account.
    # It's more or less a global needs-update flag, that causes finalization code to execute.
    touch "$NEEDS_UPDATE_FLAG_FILE"
fi

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" ]]; then
    ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"
fi
