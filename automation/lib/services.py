#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

from lib.fetch import fetch_url
from lib.networks import Network
import json
import sys


# Public location of the services.json file.
SERVICES_URL = "https://s3.us-east-2.amazonaws.com/ndau-json/services_localnet.json"


def fetch_services():
    """
    Fetch and return the json text from services.json.
    """

    services_response = fetch_url(SERVICES_URL)
    if services_response is None:
        sys.exit(f"Unable to fetch {SERVICES_URL}")

    return services_response.content


def parse_all_services(services_json):
    """
    Return API and RPC dictionaries with network name keys and values from parse_services().
    """

    network_apis = {}
    network_rpcs = {}

    for network in list(Network):
        network_name = str(network)
        network_apis[network_name], network_rpcs[network_name] = parse_services(
            network_name, None, services_json
        )

    return network_apis, network_rpcs


def parse_services(network_name, node_name, services_json):
    """
    Return a dictionary with node name keys and protocol://domain:port for the values for the
    given network's node from services_json.
    The dictionary will contain all urls on the given network if node_name is None.
    Two dictionaires are returned, one with ndauapi urls, one with tendermint RPC urls.
    """

    # Key names in services.json
    networks_name = "networks"
    nodes_name = "nodes"
    api_name = "api"
    rpc_name = "rpc"

    try:
        services_obj = json.loads(services_json)
    except:
        services_obj = None
    if services_obj is None:
        sys.exit("Unable to parse services json")

    if not networks_name in services_obj:
        sys.exit("Unable to find networks object")
    networks_obj = services_obj[networks_name]

    if not network_name in networks_obj:
        sys.exit(f"Unable to find {network_name} network object")
    network_obj = networks_obj[network_name]

    if not nodes_name in network_obj:
        sys.exit("Unable to find nodes object")
    nodes_obj = network_obj[nodes_name]

    # Ensure support for testnet-backup and mainnet-backup.  They are not published in
    # services.json, but we know of their existence and we want to manage them on AWS.
    if network_name == "testnet" or network_name == "mainnet":
        new_node_name = f"{network_name}-backup"
        if not new_node_name in nodes_obj:
            nodes_obj[new_node_name] = {
                api_name: f"{new_node_name}.ndau.tech:3030",
                rpc_name: f"{new_node_name}.ndau.tech:26670",
            }

    apis = {}
    rpcs = {}

    for node_obj_name in nodes_obj:
        if node_name is None or node_name == node_obj_name:
            node_obj = nodes_obj[node_obj_name]
            if not api_name in node_obj:
                sys.exit(f"Unable to find api object in {node_obj}")
            if not rpc_name in node_obj:
                sys.exit(f"Unable to find rpc object in {node_obj}")
            if network_name == "localnet" or network_name == "testnet":
                apis[node_obj_name] = f"http://{node_obj[api_name]}"
                rpcs[node_obj_name] = f"http://{node_obj[rpc_name]}"
            else:
                apis[node_obj_name] = f"https://{node_obj[api_name]}"
                rpcs[node_obj_name] = f"https://{node_obj[rpc_name]}"

    return apis, rpcs
