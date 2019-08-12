#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_health(url):
    """
    Get the health of the node at the given API url.
    """

    response = fetch_url(f"{url}/health")

    try:
        health = response.content.decode("utf-8").strip('"').rstrip('"\n')
        if len(health) != 0:
            return health
        # Blank health isn't really "healthy", but there was a response, so it's not "bad" either.
        return "ILL"
    except:
        return "BAD"


def main():
    """
    Print the health for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    health = get_health(url)
    print(health)


if __name__ == "__main__":
    main()
