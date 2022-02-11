#!/bin/bash

CMDBIN_DIR="$(go env GOPATH)/src/github.com/ndau/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

SERVICES_URL="https://s3.us-east-2.amazonaws.com/ndau-json/services.json"
NETWORK="mainnet"


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
# ARGS=("$@")
# for arg in "${ARGS[@]}"; do
while test $# -gt 0
do
    echo "arg = $1"
    arg=$1
    shift
    if [ "$arg" = "--needs-update" ]; then
        NEEDS_UPDATE=1
    fi
    if [[ "$arg" = "--update-default-ndauhome" || "$arg" = "-U" ]]; then
        UPDATE_DEFAULT_NDAUHOME=1
    fi
    if [ "$arg" = "--snapshot" ]; then
        SNAPSHOT_NAME=$1
        echo "snapshot name = $SNAPSHOT_NAME"
    fi
    if [ "$arg" = "--node" ]; then
        NODE_NUM=$1
        echo "node num = $NODE_NUM"
    fi
    if [ "$arg" = "--verifier" ]; then
        VERIFIER=1
        echo "node is VERIFIER"
    fi
done

copy_snapshot() {
    node_num="$1"
    echo copying snapshot for node "ndau-$node_num"

    if [ -z "$VERIFIER" ]; then
        echo extracting identity
        cat "$CMDBIN_DIR/NODE_ID-$node_num.b64" | base64 -D | tar xfvz -
    fi
    cp -r "$SNAPSHOT_NOMS_DATA_DIR" "$NOMS_NDAU_DATA_DIR-$node_num"
    cp -r "$SNAPSHOT_REDIS_DATA_DIR" "$REDIS_NDAU_DATA_DIR-$node_num"
    cp -r "$SNAPSHOT_TENDERMINT_HOME_DIR" "$TENDERMINT_NDAU_DATA_DIR-$node_num"
    # Tendermint complains if this file isn't here, but it can be empty json.
    pvs_dir="$TENDERMINT_NDAU_DATA_DIR-$node_num/data"
    pvs_file="$pvs_dir/priv_validator_state.json"
    if [ ! -f "$pvs_file" ]; then
        mkdir -p "$pvs_dir"
        echo "{}" > "$pvs_file"
    fi
}

config_tm() {
    node_num="$1"
    tm_ndau_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"

    ./build/tendermint init --home "$tm_ndau_home"

    sed -i '' -E \
        -e 's/^(create_empty_blocks =) (.*)/\1 false/' \
        -e 's/^(addr_book_strict =) (.*)/\1 false/' \
        -e 's/^(pex =) (.*)/\1 '"$PEX"'/' \
        -e 's/^(seeds =) (.*)/\1 "'"$SEEDS"'"/' \
        -e 's/^(seed_mode =) (.*)/\1 '"$SEED_MODE"'/' \
        -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
        -e 's/^(log_format =) (.*)/\1 "json"/' \
        -e 's/^(log_level =) (.*)/\1 "'"$TM_LOG_LEVEL"'"/' \
        -e 's/^(timeout_prevote =) (.*)/\1 "3s"/' \
        -e 's/^(timeout_precommit =) (.*)/\1 "3s"/' \
        -e 's/^(timeout_commit =) (.*)/\1 "3s"/' \
        -e 's/^(timeout_broadcast_tx_commit =) (.*)/\1 "30s"/' \
        -e 's/^(moniker =) (.*)/\1 \"'"$MONIKER_PREFIX"'-'"$node_num"'\"/' \
        "$tm_ndau_home/config/config.toml"
}

config_ndau() {
    node_num="$1"
    ndau_home="$NODE_DATA_DIR-$node_num"
    ndau_rpc_port=$((TM_RPC_PORT + node_num))
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    NDAUHOME="$ndau_home" ./ndau conf "$ndau_rpc_addr"

    # if the node configuration file does not exist or it does not contain
    # the node reward webhook key, then inject that key into the file
    nrw="NodeRewardWebhook"
    confpath="$ndau_home/ndau/config.toml"
    if [ -f "$confpath" ]; then
        confjs=$(toml2json "$confpath")
    else
        confjs="{}"
    fi
    if [ -z "$(jq ".$nrw // empty" <(echo $confjs))" ]; then
        confjs=$(jq -c ". + {\"$nrw\": \"http://localhost:3000/claim_winner\"}" <(echo $confjs))
    fi
    echo "$confjs" | json2toml > "$confpath"
    if [ ! -z "$SNAPSHOT_NAME" ]; then
        cat "$CMDBIN_DIR/../docker/image/docker-config-mainnet.toml" >> "$confpath"
    else
        cat "$CMDBIN_DIR/../docker/image/docker-config-default.toml" >> "$confpath"
    fi
}

set_peers_and_validators() {
    node_num="$1"

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
        jq ".validators = []" "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
            mv "$tm_ndau_genesis.new" "$tm_ndau_genesis"

        # Use something better than "test-chain-..." for the chain_id.
        jq ".chain_id=\"$CHAIN_ID\"" "$tm_ndau_genesis" > "$tm_ndau_genesis.new" && \
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
}

if [ ! -z "$SNAPSHOT_NAME" ]; then
    SNAPSHOT_DIR="$CMDBIN_DIR/snapshot"
    if [ ! -d "$SNAPSHOT_DIR" ]; then
        mkdir -p "$SNAPSHOT_DIR"
        SNAPSHOT_FILE="$SNAPSHOT_NAME.tgz"
        SNAPSHOT_URL="https://s3.amazonaws.com"
        SNAPSHOT_BUCKET="ndau-snapshots"

        echo "Fetching $SNAPSHOT_FILE..."
        curl -s -o "$SNAPSHOT_DIR/$SNAPSHOT_FILE" "$SNAPSHOT_URL/$SNAPSHOT_BUCKET/$SNAPSHOT_FILE"

        echo "Extracting $SNAPSHOT_FILE..."
        cd "$SNAPSHOT_DIR" || exit 1
        tar -xf "$SNAPSHOT_FILE"
    fi

    echo "Validating $SNAPSHOT_DIR..."
    if [ ! -d "$SNAPSHOT_DIR" ]; then
        echo "Could not find snapshot directory: $SNAPSHOT_DIR"
        exit 1
    fi
    SNAPSHOT_DATA_DIR="$SNAPSHOT_DIR/data"
    if [ ! -d "$SNAPSHOT_DATA_DIR" ]; then
        echo "Could not find data directory: $SNAPSHOT_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_NOMS_DATA_DIR="$SNAPSHOT_DATA_DIR/noms"
    if [ ! -d "$SNAPSHOT_NOMS_DATA_DIR" ]; then
        echo "Could not find noms data directory: $SNAPSHOT_NOMS_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_REDIS_DATA_DIR="$SNAPSHOT_DATA_DIR/redis"
    if [ ! -d "$SNAPSHOT_REDIS_DATA_DIR" ]; then
        echo "Could not find redis data directory: $SNAPSHOT_REDIS_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_HOME_DIR="$SNAPSHOT_DATA_DIR/tendermint"
    if [ ! -d "$SNAPSHOT_TENDERMINT_HOME_DIR" ]; then
        echo "Could not find tendermint home directory: $SNAPSHOT_TENDERMINT_HOME_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_CONFIG_DIR="$SNAPSHOT_TENDERMINT_HOME_DIR/config"
    if [ ! -d "$SNAPSHOT_TENDERMINT_CONFIG_DIR" ]; then
        echo "Could not find tendermint config directory: $SNAPSHOT_TENDERMINT_CONFIG_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_GENESIS_FILE="$SNAPSHOT_TENDERMINT_CONFIG_DIR/genesis.json"
    if [ ! -f "$SNAPSHOT_TENDERMINT_GENESIS_FILE" ]; then
        echo "Could not find tendermint genesis file: $SNAPSHOT_TENDERMINT_GENESIS_FILE"
        exit 1
    fi


    # Move the snapshot data dir where applications expect it, then remove the temp snapshot dir.
#    mv "$SNAPSHOT_DATA_DIR" "$DATA_DIR"
    pushd $SNAPSHOT_DATA_DIR
    if [ -z "$NODE_NUM" ]; then
        for node_num in $(seq 0 "$HIGH_NODE_NUM");
        do
            copy_snapshot "$node_num"
        done
    else
        copy_snapshot "$NODE_NUM"
    fi
    popd

#    rm -rf $SNAPSHOT_DIR
fi

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

# The timeout flag on linux differs from mac.
if [[ "$OSTYPE" == *"darwin"* ]]; then
    # Use -G on macOS; there is no -G option on linux.
    NC_TIMEOUT_FLAG="-G"
else
    # Use -w on linux; the -w option does not work on macOS.
    NC_TIMEOUT_FLAG="-w"
fi


test_peer() {
    ip="$1"
    port="$2"

    if [ -z "$ip" ] || [ -z "$port" ]; then
        echo "Missing p2p ip or port: ip=($ip) port=($port)"
        exit 1
    fi

    echo "Testing connection to peer $ip:$port..."
    if ! nc "$NC_TIMEOUT_FLAG" 5 -z "$ip" "$port"; then
        echo "Could not reach peer"
        exit 1
    fi
}

get_peer_id() {
    protocol="$1"
    ip="$2"
    port="$3"

    if [ -z "$protocol" ] || [ -z "$ip" ] || [ -z "$port" ]; then
        echo "Missing rpc protocol, ip or port: protocol=($protocol) ip=($ip) port=($port)"
        exit 1
    fi

    url="$protocol://$ip:$port"
    echo "Getting peer info for $url..."
    PEER_ID=$(curl -s --connect-timeout 5 "$url/status" | jq -r .result.node_info.id)
    if [ -z "$PEER_ID" ]; then
        echo "Could not get peer id"
        exit 1
    fi
    echo "Peer id: $PEER_ID"
}

NODE_ID="localnode"

# NODE_NUM is which mainnet node we are.  It will be blank if we're not named "mainnet-N".
# This also works for devnet, testnet and any "othernet" with nodes named appropriately.
NODE_NUM=$(echo "$NODE_ID" | sed -n -e 's/^[a-z]\{3,\}net-\([0-9]\{1\}\)$/\1/p')


# Return an index in [0,len].  Choose randomly if we're not one of the initial mainnet nodes.
# Otherwise, return an index that ensures this node will be referenced by the other peers via
# different IPs.  This guards against the pathological case of all peers referring to a given
# node at one IP, which would cause the peers to lose contact with that node if its IP changes.
choose_ip_index() {
    ips_len="$1"
    peer_idx="$2"

    # If we're not a mainnet node, or if the node num is invalid, choose the IP index randomly.
    # We only want to do the IP-cycling logic for the initial 5 nodes of a network.  Anything
    # outside of that is out of our control and we can't rely on the peer count and ordering.
    if [ -z "$NODE_NUM" ] || [ "$NODE_NUM" -ge 5 ]; then
        echo $((RANDOM % ips_len))
    else
        # On mainnet, the peers list has a length of one less than the number of nodes.
        # Assuming we have two IPs per peer, here is what each node will use for their peers' IPs:
        #   mainnet-0: ___, IP0, IP1, IP0, IP1
        #   mainnet-1: IP1, ___, IP0, IP1, IP0
        #   mainnet-2: IP0, IP1, ___, IP0, IP1
        #   mainnet-3: IP1, IP0, IP1, ___, IP0
        #   mainnet-4: IP0, IP1, IP0, IP1, ___
        # So for each node, half of its peers will connect to it via IP0, the other half via IP1.
        # And here's how it'd work if there were three IPs to choose from:
        #   mainnet-0: ___, IP0, IP1, IP2, IP0
        #   mainnet-1: IP1, ___, IP2, IP0, IP1
        #   mainnet-2: IP2, IP0, ___, IP1, IP2
        #   mainnet-3: IP0, IP1, IP2, ___, IP0
        #   mainnet-4: IP1, IP2, IP0, IP1, ___
        # And here's how it'd work on devnet (where every node has all 5 peers including itself):
        #   devnet-0:  ___, IP1, IP2, IP0, IP1
        #   devnet-1:  IP1, ___, IP0, IP1, IP2
        #   devnet-2:  IP2, IP0, ___, IP2, IP0
        #   devnet-3:  IP0, IP1, IP2, ___, IP1
        #   devnet-4:  IP1, IP2, IP0, IP1, ___
        # The important part is that we use many different IPs per column (a given peer).
        # For devnet, there is a slight weakness that only 2/3 of IPs are used for devnet-2.
        echo $(((NODE_NUM + peer_idx) % ips_len))
    fi
}

echo Configuring tendermint...
cd "$TENDERMINT_DIR" || exit 1

if [ -z "$NODE_NUM" ]; then
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        config_tm "$node_num"
    done
else
    config_tm "$NODE_NUM"
fi

if [ ! -z "$VERIFIER" ]; then
    echo "Fetching $SERVICES_URL..."
    services_json=$(curl -s "$SERVICES_URL")
    # shellcheck disable=SC2207
    # it works well enough for now
    p2ps=($(echo "$services_json" | jq -r ".networks.$NETWORK.nodes[].p2p"))
    # shellcheck disable=SC2207
    # it works well enough for now
    rpcs=($(echo "$services_json" | jq -r ".networks.$NETWORK.nodes[].rpc"))

    len="${#rpcs[@]}"
    if [ "$len" = 0 ]; then
        echo "No nodes published for network: $NETWORK"
        exit 1
    fi

    # The RPC connections must be made through https.
    for node in $(seq 0 $((len - 1))); do
        rpcs[$node]="https://${rpcs[$node]}"
    done

    PEERS_P2P=$(join_by , "${p2ps[@]}")
    PEERS_RPC=$(join_by , "${rpcs[@]}")
    persistent_peers=()
    IFS=',' read -ra peers_p2p <<< "$PEERS_P2P"
    IFS=',' read -ra peers_rpc <<< "$PEERS_RPC"
    len="${#peers_p2p[@]}"
    if [ "$len" != "${#peers_rpc[@]}" ]; then
        echo "The length of P2P and RPC peers must match"
        exit 1
    fi
    if [ "$len" -gt 0 ]; then
        for peer in $(seq 0 $((len - 1))); do
            IFS=':' read -ra pieces <<< "${peers_p2p[$peer]}"
            p2p_ip="${pieces[0]}"
            p2p_port="${pieces[1]}"

            test_peer "$p2p_ip" "$p2p_port"

            IFS=':' read -ra pieces <<< "${peers_rpc[$peer]}"
            rpc_protocol="${pieces[0]}"
            rpc_ip="${pieces[1]}"
            rpc_port="${pieces[2]}"

            # Since we split on colon, the double-slash is stuck to the ip.  Remove it.
            rpc_ip="${rpc_ip:2}"

            PEER_ID=""
            get_peer_id "$rpc_protocol" "$rpc_ip" "$rpc_port"
            persistent_peers+=("$PEER_ID@$p2p_ip:$p2p_port")
        done
    fi
    PERSISTENT_PEERS=$(join_by , "${persistent_peers[@]}")
    echo "Persistent peers: '$PERSISTENT_PEERS'"
    persistent_peers_new=()
    peer_idx=0
    for peer in "${persistent_peers[@]}"; do
        # Get the id and domain surrounding the '@'.
        IFS='@' read -ra split <<< "$peer"
        peer_id="${split[0]}"
        host_and_port="${split[1]}"

        # separate the host and port, delimited by `:`. e.g. `something:3000` or `127.0.0.1:4242`
        IFS=':' read -ra split <<< "$host_and_port"
        ip_or_domain="${split[0]}"
        peer_port="${split[1]}"

        # If it's already an ip, leave it as is.  Otherwise, convert it from a domain name to an ip.
        if [[ "$ip_or_domain" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
            peer_ip="$ip_or_domain"
        else
            domain="$ip_or_domain"

            # A sed-friendly whitespace pattern: space and tab.
            WHITE="[ 	]"
            # Look for "...A...<IP>".
            ips=($(dig +noall +answer "$domain" | \
                    sed -n -e 's|^.*'"$WHITE"'\{1,\}A'"$WHITE"'\{1,\}\(.*\)$|\1|p'))
            # Sort for a well-defined order.
            IFS=$'\n' ips=($(sort <<<"${ips[*]}"))

            ips_len="${#ips[@]}"
            if [ "$ips_len" = 0 ]; then
                peer_ip=""
                echo "WARNING: Unable to find IP for $domain; skipping peer $peer"
            else
                # Choose an IP.  All A records are assumed to be valid.  That's their purpose.
                ip_idx=$(choose_ip_index $ips_len $peer_idx)
                peer_ip="${ips[$ip_idx]}"
                echo "Using IP $peer_ip for peer $peer"
            fi
        fi

        # We only keep peers for which valid IPs were found.
        if [ ! -z "$peer_ip" ]; then
            persistent_peers_new+=("$peer_id@$peer_ip:$peer_port")
        fi
        peer_idx=$((peer_idx + 1))
    done

    PERSISTENT_PEERS_WITH_IPS=$(join_by , "${persistent_peers_new[@]}")

    echo "Persistent peers new: '$PERSISTENT_PEERS_WITH_IPS'"

    sed -i -E \
        -e 's|^\(persistent_peers =\) \(.*\)|\1 "'"$PERSISTENT_PEERS_WITH_IPS"'"|' \
        "$TENDERMINT_NDAU_DATA_DIR-0/config/config.toml"
fi

# Point tendermint nodes to each other if there are more than one node in the localnet.
if [[ "$NODE_COUNT" -gt 1 ]]; then
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

        peer_id=$(./build/tendermint show_node_id --home "$tm_ndau_home")
        peer_port=$((TM_P2P_PORT + node_num))
        peer="$peer_id@127.0.0.1:$peer_port"
        ndau_peers+=("$peer")

        ndau_addresses+=("$(jq -c .address "$tm_ndau_priv")")
        ndau_pub_keys+=("$(jq -c .pub_key "$tm_ndau_priv")")
    done

    # Share the peer list with every node (minus each node's own peer id).
    if [ -z "$NODE_NUM" ]; then
        for node_num in $(seq 0 "$HIGH_NODE_NUM");
        do
            set_peers_and_validators "$node_num"
        done
    else
        set_peers_and_validators "$NODE_NUM"
    fi
fi

echo Configuring ndau...
cd "$COMMANDS_DIR" || exit 1

if [ -z "$NODE_NUM" ]; then
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        config_ndau "$node_num"
    done
else
    config_ndau "$NODE_NUM"
fi

# Make sure the genesis files exist, since steps after this require them.
# The system accounts toml is optional.
if [[ ! -f "$SYSTEM_VARS_TOML"  && -z "$SNAPSHOT_NAME" ]]; then
    mkdir -p "$GENESIS_FILES_DIR"
    ./generate -v -g "$SYSTEM_VARS_TOML" -a "$SYSTEM_ACCOUNTS_TOML"
fi

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" ]]; then
    node_num=0
    ndau_rpc_port=$((TM_RPC_PORT + node_num))
    ndau_rpc_addr="http://localhost:$ndau_rpc_port"

    ./ndau conf "$ndau_rpc_addr"
    if [ -f "$SYSTEM_ACCOUNTS_TOML" ]; then
        ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"
    fi
fi

# Use this as a flag for run.sh to know whether to update ndau conf and chain with the
# genesis files, etc.

if [ "$NEEDS_UPDATE" != 0 ]; then
    for node_num in $(seq 0 "$HIGH_NODE_NUM");
    do
        ndau_home="$NODE_DATA_DIR-$node_num"

        if [ -f "$SYSTEM_ACCOUNTS_TOML" ]; then
            NDAUHOME="$ndau_home" ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"
        fi

        # Generate noms data for ndau node 0, copy from node 0 otherwise.
        data_dir="$NOMS_NDAU_DATA_DIR-$node_num"
        if [ "$node_num" = 0 ]; then
            if [ -f "$SYSTEM_ACCOUNTS_TOML" ]; then
                NDAUHOME="$ndau_home" \
                ./ndaunode -use-ndauhome \
                           -genesisfile "$SYSTEM_VARS_TOML" \
                           -asscfile "$SYSTEM_ACCOUNTS_TOML"
            else
                NDAUHOME="$ndau_home" \
                ./ndaunode -use-ndauhome \
                           -genesisfile "$SYSTEM_VARS_TOML"
            fi
            mv "$ndau_home/ndau/noms" "$data_dir"

            # set var below if ETL step is to be run. this needs to be here because ETL needs to
            # push data direct to noms dir and before noms starts
            if [ "$RUN_ETL" = "1" ]; then
                "$CMDBIN_DIR"/etl.sh "$node_num"
            fi
        else
            echo "  copying ndau noms from node 0 to node $node_num"
            cp -r "$NOMS_NDAU_DATA_DIR-0" "$data_dir"
        fi
    done

fi

if [[ "$UPDATE_DEFAULT_NDAUHOME" != "0" && -f "$SYSTEM_ACCOUNTS_TOML" ]]; then
    ./ndau conf update-from "$SYSTEM_ACCOUNTS_TOML"
fi

