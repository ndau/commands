#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url
import json


def get_height(url):
    """
    Get the current height of the node at the given api url.
    """

    # Key names in response json.
    block_meta_name = "block_meta"
    header_name = "header"
    height_name = "height"

    response = fetch_url(f"{url}/block/current")

    if not response is None:
        try:
            block_obj = json.loads(response.content)
        except:
            block_obj = None
        if not block_obj is None and block_meta_name in block_obj:
            block_meta_obj = block_obj[block_meta_name]
            if not block_meta_obj is None and header_name in block_meta_obj:
                header_obj = block_meta_obj[header_name]
                if not header_obj is None and height_name in header_obj:
                    return header_obj[height_name]

    return 0


def main():
    """
    Print the height for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    height = get_height(url)
    print(height)


if __name__ == '__main__':
    main()
