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
        info_obj = json.loads(response.content)
        if not info_obj is None and result_name in info_obj:
            result_obj = info_obj[result_name]
            if not result_obj is None and peers_name in result_obj:
                return result_obj[peers_name]

    return 0


def main():
    """
    Print the peer count for the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    peers = get_peers(url)
    print(peers)


if __name__ == '__main__':
    main()
