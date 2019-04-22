#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

# The PERSISTENT_PEERS may contain domain names.  Convert them to IPs.
# Peers are comma-separated and each peer is of the form "id@ip_or_domain_name:port".
persistent_peers=()
IFS=',' read -ra peers <<< "$PERSISTENT_PEERS"
for peer in "${peers[@]}"; do echo "$peer"; done
    # Get the port after the ':'.
    IFS=':' read -ra split <<< "$peer"
    peer_port="${split[1]}"

    # Get the id and domain surrounding the '@'.
    # The peer id will have a double-slash prefix, but it just goes along for the ride.
    IFS='@' read -ra pieces <<< "$peer"
    peer_id="${pieces[0]}"
    ip_or_domain="${pieces[1]}"

    # If it's already an ip, leave it as is.  Otherwise, convert it from a domain name to an ip.
    if [[ "$ip_or_domain" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        peer_ip="$ip_or_domain"
    else
        # A sed-friendly whitespace pattern: space and tab.
        WHITE="[ 	]"
        # Look for "...A...<IP>".
        ips=($(dig +noall +answer "$ip_or_domain" | \
                   sed -n -e 's|^.*'"$WHITE"'\{1,\}A'"$WHITE"'\{1,\}\(.*\)|\1|p'))
        len="${#ips[@]}"
        if [ "$len" = 0 ]; then
            peer_ip=""
            echo "WARNING: Unable to find IP for $ip_or_domain; skipping peer $peer"
        else
            # Choose one at random.  All A records are assumed to be valid.  That's their purpose.
            index=$((RANDOM % len))
            peer_ip="${ips[$index]}"
            echo "Using IP $peer_ip for peer $peer"
        fi
    fi

    # We only keep peers for which valid IPs were found.
    if [ ! -z "$peer_ip" ]; then
        # The peer id already has the double-slash in front of it.
        persistent_peers+=("$peer_id@$peer_ip:$peer_port")
    fi
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

PERSISTENT_PEERS_WITH_IPS=$(join_by , "${persistent_peers[@]}")

sed -i -E \
    -e 's|^(persistent_peers =) (.*)|\1 "'"$PERSISTENT_PEERS_WITH_IPS"'"|' \
    "$TM_DATA_DIR/config/config.toml"
