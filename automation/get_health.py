#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_health(url):
    """
    Get the health of the node at the given api url.
    """

    response = fetch_url(f"{url}/health")

    if not response is None:
        health_content = response.content
        if not health_content is None:
            return health_content.decode("utf-8").strip('"').rstrip('"\n')

    return "BAD"


def main():
    """
    Print the health for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    health = get_health(url)
    print(health)


if __name__ == '__main__':
    main()
