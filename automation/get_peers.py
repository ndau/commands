#!/usr/bin/env python3

from lib.args import get_args
from lib.services import get_network_urls
import json
import requests


def main():
    """
    Return peer count json for nodes on a network.
    """

    network_name, node_name = get_args()

    apis, rpcs = get_network_urls(network_name, node_name)

    peers = {}

    # Key names in response json.
    result_name = "result"
    peers_name = "n_peers"

    for network in rpcs:
        url = rpcs[network]
        response = requests.get(f"{url}/net_info")
        num_peers = 0
        if not response is None:
            info_obj = json.loads(response.content)
            if not info_obj is None and result_name in info_obj:
                result_obj = info_obj[result_name]
                if not result_obj is None and peers_name in result_obj:
                    num_peers = result_obj[peers_name]
        peers[network] = num_peers

    print(json.dumps(peers))


if __name__ == '__main__':
    main()
