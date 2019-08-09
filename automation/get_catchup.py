#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_catchup(url):
    """
    Get the current catch up status of the node at the given RPC url.
    """

    response = fetch_url(f"{url}/status")

    try:
        sync_info = response.json()["result"]["sync_info"]
        if sync_info["catching_up"] or int(sync_info["latest_block_height"]) == 0:
            return "CATCHINGUP"
        return "COMPLETE"
    except:
        return "UNKNOWN"


def main():
    """
    Print the current catch up status of the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    catchup = get_catchup(url)
    print(catchup)


if __name__ == "__main__":
    main()
