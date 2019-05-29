#!/usr/bin/env python3

from lib.args import get_args
from lib.services import get_services
import json
import requests


def main():
    """
    Dump health json for nodes on a network.
    """

    network_name, node_name = get_args()

    apis, rpcs = get_services(network_name, node_name)

    healths = {}

    for network in apis:
        url = apis[network]
        response = requests.get(f"{url}/health")
        health = "BAD"
        if not response is None:
            health_content = response.content
            if not health_content is None:
                health = health_content.decode("utf-8").strip('"').rstrip('"\n')
        healths[network] = health

    print(json.dumps(healths))


if __name__ == '__main__':
    main()
