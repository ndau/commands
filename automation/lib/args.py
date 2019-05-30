#!/usr/bin/env python3

from lib.networks import Network
from argparse import ArgumentParser
from enum import Enum


class UrlKind(Enum):
    """
    Supported url kinds.
    """

    API = 'API'
    RPC = 'RPC'

    def __str__(self):
        return self.value


def get_url(kind):
    """
    Return the url argument.
    """

    parser = ArgumentParser()
    parser.add_argument(f"url", help=f"{kind} URL of the form protocol://domain:port")
    args = parser.parse_args()
    return args.url


def get_net():
    """
    Return the net argument.
    """

    parser = ArgumentParser()
    parser.add_argument(
        'net', nargs="?", type=Network, choices=list(Network), help='network name')
    args = parser.parse_args()
    return args.net
