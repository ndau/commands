#!/usr/bin/env python3

from lib.args import get_args
from lib.services import get_network_urls
import json
import requests


def main():
    """
    Return SHA json for nodes on a network.
    """

    network_name, node_name = get_args()

    apis, rpcs = get_network_urls(network_name, node_name)

    shas = {}

    for network in apis:
        url = apis[network]
        response = requests.get(f"{url}/version")
        shas[network] = "UNKNOWN"
        if not response is None:
            sha_name = "NdauSha"
            version_obj = json.loads(response.content)
            if not version_obj is None and sha_name in version_obj:
                shas[network] = version_obj[sha_name]

    print(json.dumps(shas))


if __name__ == '__main__':
    main()
