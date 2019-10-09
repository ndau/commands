#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

from get_catchup import get_catchup
from get_health import get_health
from get_height import get_height
from get_peers import get_peers
from get_sha import get_sha
from get_version import get_version
from lib.args import get_net
from lib.networks import Network
from lib.services import fetch_services, parse_all_services
import os
import sys


def print_at(x, y, text):
    """
    Print the given text at the given shell position.
    (1,1) is the upper left position with positive x to the right and positive y downward.
    """

    sys.stdout.write(f"\x1b7\x1b[{y};{x}f{text}\x1b8")
    sys.stdout.flush()


def print_node_info(x, y, info_func, urls):
    """
    Print the given info function's return value at the given position.
    """

    for node_name in sorted(urls):
        url = urls[node_name]
        print_at(x, y, "...       ")
        result = info_func(url)
        print_at(x, y, "          ")
        print_at(x, y, result)
        y += 1


def render_hud():
    """
    Render node status on the screen until interrupted.
    """

    network = get_net()
    if network is None:
        networks = list(Network)
    else:
        networks = [network]

    os.system("clear")

    column_width = 11
    x_node = 1
    x_health = x_node + column_width + 4
    x_version = x_health + column_width
    x_sha = x_version + column_width
    x_catchup = x_sha + column_width
    x_height = x_catchup + column_width
    x_peers = x_height + column_width

    y_network = 2
    print_at(x_node, y_network, "Node")
    print_at(x_health, y_network, "Health")
    print_at(x_version, y_network, "Version")
    print_at(x_sha, y_network, "SHA")
    print_at(x_catchup, y_network, "Catchup")
    print_at(x_height, y_network, "Height")
    print_at(x_peers, y_network, "Peers")
    print_at(
        x_node,
        y_network + 1,
        "-------------- ---------- ---------- ---------- ---------- ---------- ----------",
    )

    # Fetch the api and rpc urls once.
    network_apis, network_rpcs = parse_all_services(fetch_services())

    # Render status forever.
    while True:
        y_network = 4
        for network in networks:
            network_name = str(network)
            apis = network_apis[network_name]
            rpcs = network_rpcs[network_name]

            y = y_network
            for node_name in sorted(apis):
                print_at(x_node, y, node_name)
                y += 1

            # We poll the same thing for each node in the network, rather than all things for
            # each node.  That way we don't hit a single node hard with back-to-back requests.
            print_node_info(x_health, y_network, get_health, apis)
            print_node_info(x_version, y_network, get_version, apis)
            print_node_info(x_sha, y_network, get_sha, apis)
            print_node_info(x_catchup, y_network, get_catchup, rpcs)
            print_node_info(x_height, y_network, get_height, apis)
            print_node_info(x_peers, y_network, get_peers, rpcs)

            y_network += len(apis) + 1


def main():
    """
    Display status information for all Oneiro nodes and networks.
    """

    try:
        render_hud()
    except KeyboardInterrupt:
        sys.exit(0)


if __name__ == "__main__":
    main()
