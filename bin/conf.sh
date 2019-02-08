#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Protection against conf.sh being run multiple times.
# We only want to flag for needs-update if we're being called from setup.sh or reset.sh.
NEEDS_UPDATE=0

# By default, this script only updates the chaos and ndau node configuration
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
    tm_chaos_home="$TENDERMINT_CHAOS_DATA_DIR-$node_num"
    tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"

    ./tendermint init --home "$tm_chaos_home"
    ./tendermint init --home "$tm_ndau_home"

    sed -i '' -E \
        -e 's/^(create_empty_blocks = .*)/# \1/' \
        -e 's/^(create_empty_blocks_interval =) (.*)/\1 "300s"/' \
        -e 's/^(addr_book_strict =) (.*)/\1 false/' \
        -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
        -e 's/^(moniker =) (.*)/\1 \"localnet-'"$node_num"'\"/' \
        "$tm_chaos_home/config/config.toml" \
        "$tm_ndau_home/config/config.toml"

    # Replace the test-chain-XXXX with a constant, so that peers can connect to each other.
    # Tendermint uses this chain_id as a network identifier.
    genesis_config="$TENDERMINT_CHAOS_DATA_DIR-$node_num/config/genesis"
    jq ".chain_id=\"local-chain-chaos\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"

    genesis_config="$TENDERMINT_NDAU_DATA_DIR-$node_num/config/genesis"
    jq ".chain_id=\"local-chain-ndau\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

# Point tendermint nodes to each other if there are more than one node in the localnet.
if [ "$NODE_COUNT" -gt 1 ]; then
    # Because of Tendermint's PeX feature, each node could gissip known peers to the others.
    # So for every node's config, we'd only need to tell it about one other node, not all of
    # them.  The last node therefore wouldn't need to know about any peers, because the
    # previous one will dial it up as a peer.  However, to be more like how things are done in
    # the automation repo, we share all peers with each other.
    chaos_peers=()
    ndau_peers=()
    chaos_addresses=()
    ndau_addresses=()
    chaos_pub_keys=()
    ndau_pub_keys=()

    # Build the full list of peers.
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        tm_chaos_home="$TENDERMINT_CHAOS_DATA_DIR-$node_num"
        tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"
        tm_chaos_genesis="$tm_chaos_home/config/genesis.json"
        tm_ndau_genesis="$tm_ndau_home/config/genesis.json"
        tm_chaos_priv="$tm_chaos_home/config/priv_validator_key.json"
        tm_ndau_priv="$tm_ndau_home/config/priv_validator_key.json"

        peer_id=$(./tendermint show_node_id --home "$tm_chaos_home")
        peer_port=$((TM_P2P_PORT + 2 * node_num))
        peer="$peer_id@127.0.0.1:$peer_port"
        chaos_peers+=("$peer")

        peer_id=$(./tendermint show_node_id --home "$tm_ndau_home")
        peer_port=$((TM_P2P_PORT + 2 * node_num + 1))
        peer="$peer_id@127.0.0.1:$peer_port"
        ndau_peers+=("$peer")

        chaos_addresses+=($(jq -c .address "$tm_chaos_priv"))
        ndau_addresses+=($(jq -c .address "$tm_ndau_priv"))
        chaos_pub_keys+=($(jq -c .pub_key "$tm_chaos_priv"))
        ndau_pub_keys+=($(jq -c .pub_key "$tm_ndau_priv"))

        # Clear out the validator list.  We'll reconstruct it from scratch below.
        jq ".validators = []" \
           "$tm_chaos_genesis" > "$tm_chaos_genesis.new" && \
            mv "$tm_chaos_genesis.new" "$tm_chaos_genesis"
        jq ".validators = []" \
           "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
            mv "$tm_ndau_genesis.new" "$tm_ndau_genesis"
    done

    # Share the peer list with every node (minus each node's own peer id).
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        tm_chaos_home="$TENDERMINT_CHAOS_DATA_DIR-$node_num"
        tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"
        tm_chaos_genesis="$tm_chaos_home/config/genesis.json"
        tm_ndau_genesis="$tm_ndau_home/config/genesis.json"
        tm_chaos_config="$tm_chaos_home/config/config.toml"
        tm_ndau_config="$tm_ndau_home/config/config.toml"

        non_self_peers=("${chaos_peers[@]}")
        unset 'non_self_peers[$node_num]'
        peers=$(join_by , "${non_self_peers[@]}")
        sed -i '' -E \
            -e 's/^(persistent_peers =) (.*)/\1 \"'"$peers"'\"/' \
            "$tm_chaos_config"

        non_self_peers=("${ndau_peers[@]}")
        unset 'non_self_peers[$node_num]'
        peers=$(join_by , "${non_self_peers[@]}")
        sed -i '' -E \
            -e 's/^(persistent_peers =) (.*)/\1 \"'"$peers"'\"/' \
            "$tm_ndau_config"

        # Make every node's genesis file have all nodes set up as validators.
        for peer_num in $(seq 0 "$HIGH_NODE_NUM");
        do
            a=${chaos_addresses[$peer_num]}
            k=${chaos_pub_keys[$peer_num]}
            p=10
            n="chaos-$peer_num"
            jq ".validators+=[{\"address\":$a,\"pub_key\":$k,\"power\":\"$p\",\"name\":\"$n\"}]" \
               "$tm_chaos_genesis" > "$tm_chaos_genesis.new" && \
                mv "$tm_chaos_genesis.new" "$tm_chaos_genesis"

            a=${ndau_addresses[$peer_num]}
            k=${ndau_pub_keys[$peer_num]}
            p=10
            n="ndau-$peer_num"
            jq ".validators+=[{\"address\":$a,\"pub_key\":$k,\"power\":\"$p\",\"name\":\"$n\"}]" \
               "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
                mv "$tm_ndau_genesis.new" "$tm_ndau_genesis"
        done
    done
fi

echo Configuring chaos and ndau...
cd "$COMMANDS_DIR" || exit 1

for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    ndau_home="$NODE_DATA_DIR-$node_num"
    port_offset=$((2 * node_num))
    chaos_rpc_port=$((TM_RPC_PORT + port_offset))
    ndau_rpc_port=$((TM_RPC_PORT + port_offset + 1))
    chaos_rpc_addr="http://localhost:$chaos_rpc_port"
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    NDAUHOME="$ndau_home" ./chaos conf "$chaos_rpc_addr" --ndau "$ndau_rpc_addr"
    NDAUHOME="$ndau_home" ./chaosnode --set-ndaunode "$ndau_rpc_addr"
    NDAUHOME="$ndau_home" ./ndau conf "$ndau_rpc_addr"
    NDAUHOME="$ndau_home" ./ndaunode --set-chaosnode "$chaos_rpc_addr"
done

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" ]]; then
    node_num=0
    port_offset=$((2 * node_num))
    chaos_rpc_port=$((TM_RPC_PORT + port_offset))
    ndau_rpc_port=$((TM_RPC_PORT + port_offset + 1))
    chaos_rpc_addr="http://localhost:$chaos_rpc_port"
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    ./chaos conf "$chaos_rpc_addr" --ndau "$ndau_rpc_addr"
    ./ndau conf "$ndau_rpc_addr"
fi

# Use this as a flag for run.sh to know whether to update ndau conf and chain with the
# genesis files, claim bpc account, etc.
if [ "$NEEDS_UPDATE" != 0 ]; then
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        ndau_home="$NODE_DATA_DIR-$node_num"

        ./genesis -g "$GENESIS_TOML" -n "$NOMS_CHAOS_DATA_DIR-$node_num"
        NDAUHOME="$ndau_home" ./ndau conf update-from "$ASSC_TOML"

        # For deterministic bpc account address/keys, we recover a special account with 12 eyes.
        # Since this is only for localnet/devnet/testnet (i.e. not mainnet), this is safe.
        NDAUHOME="$ndau_home" ./ndau account recover "$BPC_OPS_ACCT_NAME" \
            eye eye eye eye eye eye eye eye eye eye eye eye

        # Set up the bpc-operations identity in the chaos tool toml file.
        # Suppress the big message about next steps.
        NDAUHOME="$ndau_home" ./chaos import-assc "$SYSVAR_ID" "$ASSC_TOML" > /dev/null

        touch "$NEEDS_UPDATE_FLAG_FILE-$node_num"
    done

    # The no-node-num form of the needs-update file flags that we need to claim the bpc account.
    # It's more or less a global needs-update flag, that causes finalization code to execute.
    touch "$NEEDS_UPDATE_FLAG_FILE"
fi

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" ]]; then
    ./ndau conf update-from "$ASSC_TOML"
fi
