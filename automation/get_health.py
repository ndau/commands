#!/usr/bin/env python3

from lib.args import get_args
from lib.services import get_network_urls
import json
import requests


def main():
    """
    Return health json for nodes on a network.
    """

    network_name, node_name = get_args()

    apis, rpcs = get_network_urls(network_name, node_name)

    healths = {}

    for network in apis:
        url = apis[network]
        response = requests.get(f"{url}/health")
        if response is None:
            healths[network] = "BAD"
        else:
            healths[network] = response.content.decode("utf-8").strip('"').rstrip('"\n')

    print(json.dumps(healths))


if __name__ == '__main__':
    main()
