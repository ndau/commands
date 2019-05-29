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

    heights = {}

    # Key names in response json.
    block_meta_name = "block_meta"
    header_name = "header"
    height_name = "height"

    for network in apis:
        url = apis[network]
        response = requests.get(f"{url}/block/current")
        height = 0
        if not response is None:
            block_obj = json.loads(response.content)
            if not block_obj is None and block_meta_name in block_obj:
                block_meta_obj = block_obj[block_meta_name]
                if not block_meta_obj is None and header_name in block_meta_obj:
                    header_obj = block_meta_obj[header_name]
                    if not header_obj is None and height_name in header_obj:
                        height = header_obj[height_name]
        heights[network] = height

    print(json.dumps(heights))


if __name__ == '__main__':
    main()
