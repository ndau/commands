#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url
import json


def get_sha(url):
    """
    Get the SHA of the node at the given api url.
    """

    # Key names in response json.
    sha_name = "NdauSha"

    response = fetch_url(f"{url}/version")

    if not response is None:
        version_obj = json.loads(response.content)
        if not version_obj is None and sha_name in version_obj:
            return version_obj[sha_name]

    return "UNKNOWN"


def main():
    """
    Print the SHA for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    sha = get_sha(url)
    print(sha)


if __name__ == '__main__':
    main()
