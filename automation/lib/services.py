#!/usr/bin/env python3

from lib import constants
import json
import requests
import sys


def get_network_urls(network_name, node_name):
    """
    Return a dictionary with node name keys and protocol://domain:port for the values for the
    given network's node from services.json.
    The dictionary will contain all urls on the given network if node_name is None.
    """

    # Key names in services.json
    networks_name = "networks"
    nodes_name = "nodes"
    api_name = "api"

    services_response = requests.get(constants.SERVICES_URL)
    if services_response is None:
        sys.exit(f"Unable to fetch {constants.SERVICES_URL}")

    services_obj = json.loads(services_response.content)
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

    # Ensure support for testnet-5 and mainnet-5.  They are not published in services.json,
    # but we know of their existence and we want to manage them on AWS.
    if network_name == "testnet" or network_name == "mainnet":
        new_node_name = f"{network_name}-5"
        if not new_node_name in nodes_obj:
            nodes_obj[new_node_name] = {
                api_name: f"{new_node_name}.ndau.tech:3030"
            }

    urls = {}

    for node_obj_name in nodes_obj:
        if node_name is None or node_name == node_obj_name:
            node_obj = nodes_obj[node_obj_name]
            if not api_name in node_obj:
                sys.exit(f"Unable to find api object in {node_obj}")
            urls[node_obj_name] = f"https://{node_obj[api_name]}"

    return urls
