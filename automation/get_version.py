#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_version(url):
    """
    Get the version of the node at the given API url.
    """

    response = fetch_url(f"{url}/version")

    try:
        return response.json()["NdauVersion"]
    except:
        return "UNKNOWN"


def main():
    """
    Print the version for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    version = get_version(url)
    print(version)


if __name__ == "__main__":
    main()
