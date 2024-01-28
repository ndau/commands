#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

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

# The PERSISTENT_PEERS may contain domain names.  Convert them to IPs.
# Peers are comma-separated and each peer is of the form "id@ip_or_domain_name:port".
persistent_peers=()
IFS=',' read -ra peers <<< "$PERSISTENT_PEERS"
peer_idx=0
for peer in "${peers[@]}"; do
    # Get the id and domain surrounding the '@'.
    IFS='@' read -ra split <<< "$peer"
    peer_id="${split[0]}"
    host_and_port="${split[1]}"

    # separate the host and port, delimited by `:`. e.g. `something:3000` or `127.0.0.1:4242`
    IFS=':' read -ra split <<< "$host_and_port"
    ip_or_domain="${split[0]}"
    peer_port="${split[1]}"

    # If it's already an ip, leave it as is.  Otherwise, convert it from a domain name to an ip.
#    if [[ "$ip_or_domain" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        peer_ip="$ip_or_domain"
#    else
#        domain="$ip_or_domain"
#
#        # A sed-friendly whitespace pattern: space and tab.
#        WHITE="[ 	]"
#        # Look for "...A...<IP>".
#        ips=($(dig +noall +answer "$domain" | \
#                   sed -n -e 's|^.*'"$WHITE"'\{1,\}A'"$WHITE"'\{1,\}\(.*\)$|\1|p'))
#        # Sort for a well-defined order.
#        IFS=$'\n' ips=($(sort <<<"${ips[*]}"))
#
#        ips_len="${#ips[@]}"
#        if [ "$ips_len" = 0 ]; then
#            peer_ip=""
#            echo "WARNING: Unable to find IP for $domain; skipping peer $peer"
#        else
#            # Choose an IP.  All A records are assumed to be valid.  That's their purpose.
#            ip_idx=$(choose_ip_index $ips_len $peer_idx)
#            peer_ip="${ips[$ip_idx]}"
#            echo "Using IP $peer_ip for peer $peer"
#        fi
#    fi

    # We only keep peers for which valid IPs were found.
    if [ ! -z "$peer_ip" ]; then
        persistent_peers+=("$peer_id@$peer_ip:$peer_port")
    fi
    peer_idx=$((peer_idx + 1))
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

PERSISTENT_PEERS_WITH_IPS=$(join_by , "${persistent_peers[@]}")

sed -i -E \
    -e 's|^(persistent_peers =) (.*)|\1 "'"$PERSISTENT_PEERS_WITH_IPS"'"|' \
    "$TM_DATA_DIR/config/config.toml"
