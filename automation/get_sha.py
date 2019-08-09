#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_sha(url):
    """
    Get the SHA of the node at the given API url.
    """

    response = fetch_url(f"{url}/version")

    try:
        return response.json()["NdauSha"]
    except:
        return "UNKNOWN"


def main():
    """
    Print the SHA for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    sha = get_sha(url)
    print(sha)


if __name__ == "__main__":
    main()
