#!/usr/bin/env python3

from argparse import ArgumentParser
from enum import Enum


class Network(Enum):
    """
    Supported network names.
    """

    devnet = 'devnet'
    testnet = 'testnet'
    mainnet = 'mainnet'

    def __str__(self):
        return self.value


def get_args():
    """
    Return the network name and node name from the command line.
    If no node was specified, then "all" is assumed and None is returned for it.
    """

    parser = ArgumentParser()

    parser.add_argument('net', type=Network, choices=list(Network), help='network name')
    parser.add_argument('--node', required=False, help='node name')

    args = parser.parse_args()
    network_name = str(args.net)
    node_name = args.node

    return network_name, node_name
