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
    All commands use the same arguments.
    """

    parser = ArgumentParser()

    parser.add_argument('net', type=Network, choices=list(Network), help='network name')
    parser.add_argument('--node', required=False, help='node name')

    args = parser.parse_args()

    network_name = args.net
    if args.node is None:
        node_name = "(all)"
    else:
        node_name = args.node
    print(f"Network: {network_name}")
    print(f"Node   : {node_name}")

    return args
