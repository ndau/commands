#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Protection against conf.sh being run multiple times.
# We only want to flag for needs_update if we're being called from setup.sh or reset.sh.
NEEDS_UPDATE=0

# Process command line arguments.
ARGS=("$@")
for arg in "${ARGS[@]}"; do
    if [ "$arg" = "--needs_update" ]; then
        NEEDS_UPDATE=1
        break
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
        -e 's/^(create_empty_blocks_interval =) (.*)/\1 300/' \
        -e 's/^(addr_book_strict =) (.*)/\1 false/' \
        -e 's/^(moniker =) (.*)/\1 \"localnet-'"$node_num"'\"/' \
        "$tm_chaos_home/config/config.toml" \
        "$tm_ndau_home/config/config.toml"

    # Replace the test-chain-XXXX with a constant, so that peers can connect to each other.
    # Tendermint uses this chain_id as a network identifier.
    genesis_config="$TENDERMINT_CHAOS_DATA_DIR-$node_num/config/genesis"
    jq ".chain_id=\"local-chain-chaos\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"
    jq ".validators[0].name=\"chaos-$node_num\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"

    genesis_config="$TENDERMINT_NDAU_DATA_DIR-$node_num/config/genesis"
    jq ".chain_id=\"local-chain-ndau\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"
    jq ".validators[0].name=\"ndau-$node_num\"" \
        "$genesis_config.json" > "$genesis_config.new.json" && \
        mv "$genesis_config.new.json" "$genesis_config.json"
done

# Point tendermint nodes to each other if there are more than one node in the localnet.
if [ "$NODE_COUNT" -gt 1 ]; then
    # Because of Tendermint's PeX feature, each node will gissip known peers to the others.
    # So for every node's config, we only need to tell it about one other node, not all of
    # them.  The last node therefore doesn't need to know about any peers, because the
    # previous one will dial it up as a peer.
    for node_num in $(seq 0 $(expr "$HIGH_NODE_NUM" - 1));
    do
        peer_num=$(expr "$node_num" + 1)

        src_tm_chaos_home="$TENDERMINT_CHAOS_DATA_DIR-$peer_num"
        src_tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$peer_num"
        dst_tm_chaos_home="$TENDERMINT_CHAOS_DATA_DIR-$node_num"
        dst_tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"

        peer_id=$(./tendermint show_node_id --home "$src_tm_chaos_home" | \
                      sed -e 's/^.*honeycomb.*$//')
        peer_id=$(echo -e "$peer_id" | tr -d '[:space:]')
        peer_port=$(expr "$TM_P2P_PORT" + 2 \* "$peer_num")
        peer="$peer_id@127.0.0.1:$peer_port"
        sed -i '' -E \
            -e 's/^(persistent_peers =) (.*)/\1 \"'"$peer"'\"/' \
            "$dst_tm_chaos_home/config/config.toml"

        peer_id=$(./tendermint show_node_id --home "$src_tm_ndau_home" | \
                      sed -e 's/^.*honeycomb.*$//')
        peer_id=$(echo -e "$peer_id" | tr -d '[:space:]')
        peer_port=$(expr "$TM_P2P_PORT" + 2 \* "$peer_num" + 1)
        peer="$peer_id@127.0.0.1:$peer_port"
        sed -i '' -E \
            -e 's/^(persistent_peers =) (.*)/\1 \"'"$peer"'\"/' \
            "$dst_tm_ndau_home/config/config.toml"
    done
fi

echo Configuring chaos and ndau...
cd "$COMMANDS_DIR" || exit 1

for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    ndau_home="$NODE_DATA_DIR-$node_num"
    port_offset=$(expr 2 \* "$node_num")
    chaos_rpc_port=$(expr "$TM_RPC_PORT" + "$port_offset")
    ndau_rpc_port=$(expr "$TM_RPC_PORT" + "$port_offset" + 1)
    chaos_rpc_addr="http://localhost:$chaos_rpc_port"
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    NDAUHOME="$ndau_home" ./chaos conf "$chaos_rpc_addr"
    NDAUHOME="$ndau_home" ./chaosnode --set-ndaunode "$ndau_rpc_addr"
    NDAUHOME="$ndau_home" ./ndau conf "$ndau_rpc_addr"
done

for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    ./genesis -g "$GENESIS_TOML" -n "$NOMS_CHAOS_DATA_DIR-$node_num"
    NDAUHOME="$NODE_DATA_DIR-$node_num" ./ndau conf update-from "$ASSC_TOML"

    # Use this as a flag for run.sh to know whether to update ndau conf and chain with the
    # genesis files.
    if [ "$NEEDS_UPDATE" != 0 ]; then
        touch "$NEEDS_UPDATE_FLAG_FILE-$node_num"
    fi
done
