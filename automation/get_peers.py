#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_peers(url):
    """
    Get the peer count of the node at the given RPC url.
    """

    response = fetch_url(f"{url}/net_info")

    try:
        return response.json()["result"]["n_peers"]
    except:
        # Return an invalid peer count to signal failure.
        return -1


def main():
    """
    Print the peer count for the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    peers = get_peers(url)
    print(peers)


if __name__ == "__main__":
    main()
