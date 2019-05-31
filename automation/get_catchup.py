#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url
import json


def get_catchup(url):
    """
    Get the current catch up status of the node at the given RPC url.
    """

    # Key names in response json.
    result_name = "result"
    sync_info_name = "sync_info"
    latest_block_height_name = "latest_block_height"
    catching_up_name = "catching_up"

    response = fetch_url(f"{url}/status")

    if not response is None:
        try:
            root_obj = json.loads(response.content)
        except:
            root_obj = None
        if not root_obj is None and result_name in root_obj:
            result_obj = root_obj[result_name]
            if not result_obj is None and sync_info_name in result_obj:
                sync_info_obj = result_obj[sync_info_name]
                if not sync_info_obj is None and \
                   latest_block_height_name in sync_info_obj and \
                   catching_up_name in sync_info_obj:
                    if int(sync_info_obj[latest_block_height_name]) > 0 and \
                       not sync_info_obj[catching_up_name]:
                        return "COMPLETE"
                    return "CATCHINGUP"

    return "UNKNOWN"


def main():
    """
    Print the current catch up status of the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    catchup = get_catchup(url)
    print(catchup)


if __name__ == '__main__':
    main()
