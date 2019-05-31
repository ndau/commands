#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url
import json


def get_peers(url):
    """
    Get the peer count of the node at the given RPC url.
    """

    # Key names in response json.
    result_name = "result"
    peers_name = "n_peers"

    response = fetch_url(f"{url}/net_info")

    if not response is None:
        try:
            root_obj = json.loads(response.content)
        except:
            root_obj = None
        if not root_obj is None and result_name in root_obj:
            result_obj = root_obj[result_name]
            if not result_obj is None and peers_name in result_obj:
                return result_obj[peers_name]

    return -1 # Invalid peer count


def main():
    """
    Print the peer count for the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    peers = get_peers(url)
    print(peers)


if __name__ == '__main__':
    main()
