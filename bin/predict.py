#!/usr/bin/python

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

"""
This tool was crafted based on the numbers gathered from this issue:
  https://github.com/oneiro-ndev/ndau/issues/186

Run ./predict.py to see Usage instructions.
"""

import sys


# Constants
SECONDS_PER_MINUTE = 60
SECONDS_PER_HOUR = 60 * SECONDS_PER_MINUTE
SECONDS_PER_DAY = 24 * SECONDS_PER_HOUR
BYTES_PER_KILOBYTE = 1024
BYTES_PER_MEGABYTE = 1024 * BYTES_PER_KILOBYTE
BYTES_PER_GIGABYTE = 1024 * BYTES_PER_MEGABYTE
BYTES_PER_TERABYTE = 1024 * BYTES_PER_GIGABYTE

# Bytes per block constants per chain.
BYTES_PER_BLOCK_CHAOS_REDIS = 15
BYTES_PER_BLOCK_CHAOS_NOMS = 6600

BYTES_PER_BLOCK_NDAU_REDIS = 98
BYTES_PER_BLOCK_NDAU_NOMS = 12000

# Tendermint bytes per block is chain-independent.
BYTES_PER_BLOCK_TENDERMINT = 3100

# Tendermint commits at least one block in this many seconds.
TENDERMINT_KEEPALIVE_SECONDS = 300
TENDERMINT_MIN_BLOCKS_PER_SECOND = 1.0 / TENDERMINT_KEEPALIVE_SECONDS


class Phase:
    """
    Each phase is a block rate with the total time to run at that rate.
    """

    def __init__(self, phase_num, blocks_per_second, total_seconds):
        self.phase_num = phase_num
        self.blocks_per_second = blocks_per_second
        self.total_seconds = total_seconds

    def __str__(self):
        return 'phase {0}:\n  blocks per second = {1}\n  total seconds     = {2}'.format(
            self.phase_num, self.blocks_per_second, self.total_seconds)


def print_usage_and_die(msg):
    """
    Print usage information for this tool and exits with a non-zero exit code.
    """

    print(msg)
    print('Usage:')
    print(
        '  ./predict.py {chaos|ndau} [[rate {bps|bpm|bph|bpd} duration {s|m|h|d}]...]'
    )
    print('Example: (ndau chain, idle for 1 day, then 2 blocks per second for 4 hours)')
    print('  ./predict.py ndau 0 bpm 1 d 2 bps 4 h')

    exit(1)


def format_byte_size(size):
    """
    Return a human-readable string for the given byte size.
    """

    if size >= BYTES_PER_TERABYTE:
        return '{0:.1f} TB'.format(float(size) / BYTES_PER_TERABYTE)
    if size >= BYTES_PER_GIGABYTE:
        return '{0:.1f} GB'.format(float(size) / BYTES_PER_GIGABYTE)
    if size >= BYTES_PER_MEGABYTE:
        return '{0:.1f} MB'.format(float(size) / BYTES_PER_MEGABYTE)
    if size >= BYTES_PER_KILOBYTE:
        return '{0:.1f} KB'.format(float(size) / BYTES_PER_KILOBYTE)
    return '{0} bytes'.format(size)


def format_seconds(seconds):
    """
    Return a human-readable string for the duration.
    """

    if seconds >= SECONDS_PER_DAY:
        return '{0:.2f} days'.format(seconds / SECONDS_PER_DAY)
    if seconds >= SECONDS_PER_HOUR:
        return '{0:.2f} hours'.format(seconds / SECONDS_PER_HOUR)
    if seconds >= SECONDS_PER_MINUTE:
        return '{0:.2f} minutes'.format(seconds / SECONDS_PER_MINUTE)
    return '{0:.2f} seconds'.format(seconds)


def parse_args():
    """
    Parse the command line arguments passed in to this tool.
    Return the bytes per block for redis and noms, as well as an array of phases to run through.
    """

    arg_index = 0
    arg_count = len(sys.argv)
    if arg_count < 2:
        print_usage_and_die('Must specify a chain')

    arg_index += 1
    chain = sys.argv[arg_index]
    if chain == 'chaos':
        redis_bpb = BYTES_PER_BLOCK_CHAOS_REDIS
        noms_bpb = BYTES_PER_BLOCK_CHAOS_NOMS
    elif chain == 'ndau':
        redis_bpb = BYTES_PER_BLOCK_NDAU_REDIS
        noms_bpb = BYTES_PER_BLOCK_NDAU_NOMS
    else:
        print_usage_and_die('Unsupported chain: ' + chain)

    arg_index += 1
    phase_arg_count = arg_count - arg_index
    if phase_arg_count & 3 != 0:
        print_usage_and_die('Each phase must have 4 arguments')

    phases = []
    phase_num = 1
    while arg_index < arg_count:
        blocks_per_second = float(sys.argv[arg_index])
        if blocks_per_second < 0:
            print_usage_and_die(
                'Unsupported phase block rate: ' + blocks_per_second)
        arg_index += 1

        rate = sys.argv[arg_index]
        if rate == 'bpm':
            blocks_per_second /= SECONDS_PER_MINUTE
        elif rate == 'bph':
            blocks_per_second /= SECONDS_PER_HOUR
        elif rate == 'bpd':
            blocks_per_second /= SECONDS_PER_DAY
        elif rate != 'bps':
            print_usage_and_die('Unsupported phase block rate scale: ' + rate)
        arg_index += 1

        total_seconds = float(sys.argv[arg_index])
        if blocks_per_second < 0:
            print_usage_and_die('Unsupported phase duration: ' + total_seconds)
        arg_index += 1

        duration = sys.argv[arg_index]
        if duration == 'm':
            total_seconds *= SECONDS_PER_MINUTE
        elif duration == 'h':
            total_seconds *= SECONDS_PER_HOUR
        elif duration == 'd':
            total_seconds *= SECONDS_PER_DAY
        elif duration != 's':
            print_usage_and_die(
                'Unsupported phase duration scale: ' + duration)
        arg_index += 1

        phases.append(Phase(phase_num, blocks_per_second, total_seconds))
        phase_num += 1

    return chain, redis_bpb, noms_bpb, BYTES_PER_BLOCK_TENDERMINT, phases


def main():
    chain, redis_bpb, noms_bpb, tendermint_bpb, phases = parse_args()

    print('predicting disk usage for the {0} chain over {1} phases'.format(
        chain, len(phases)))
    print('  redis      bytes per block = {0}'.format(redis_bpb))
    print('  noms       bytes per block = {0}'.format(noms_bpb))
    print('  tendermint bytes per block = {0}'.format(tendermint_bpb))
    print('  tendermint max seconds between blocks = {0}'.format(
        TENDERMINT_KEEPALIVE_SECONDS))
    for i in range(len(phases)):
        print(phases[i])

    blocks = 0
    redis_bytes = 0
    noms_bytes = 0
    tendermint_bytes = 0
    duration = 0

    # This loop doesn't worry about any duration overlap between phases, so it's
    # inaccurate in that way.  It's minor for any reasonable prediction query.
    for phase in phases:
        block_count = int(phase.blocks_per_second * phase.total_seconds)

        redis_bytes += block_count * redis_bpb
        noms_bytes += block_count * noms_bpb

        # Compute the extra keepalive blocks created by tendermint.
        # Only tendermint data files are affected by these empty blocks.
        if phase.blocks_per_second < TENDERMINT_MIN_BLOCKS_PER_SECOND:
            if phase.blocks_per_second == 0:
                block_count += int(TENDERMINT_MIN_BLOCKS_PER_SECOND *
                                   phase.total_seconds)
            else:
                # The keepalive timer starts over after any block is created.  So we count how
                # many keepalive blocks get created in between every normal phase block.  This
                # calculation might count one extra keepalive block at the end of the phase
                # duration.  It's minor enough that we ignore that case in favor of simplicity.
                phase_seconds_per_block = 1 / phase.blocks_per_second
                extra_blocks_per_block = \
                    int(phase_seconds_per_block / TENDERMINT_KEEPALIVE_SECONDS)

                # When there is an even multiple of keepalive blocks, we don't count the
                # last one, as if a normal phase block would be created at that time.
                if phase_seconds_per_block == int(phase_seconds_per_block) and \
                   phase_seconds_per_block % TENDERMINT_KEEPALIVE_SECONDS == 0:
                    extra_blocks_per_block -= 1

                block_count += block_count * extra_blocks_per_block

        tendermint_bytes += block_count * tendermint_bpb

        blocks += block_count
        duration += phase.total_seconds

    print('results:')
    print('  height     = {0:,} blocks'.format(blocks))
    print('  redis      = {0}'.format(format_byte_size(redis_bytes)))
    print('  noms       = {0}'.format(format_byte_size(noms_bytes)))
    print('  tendermint = {0}'.format(format_byte_size(tendermint_bytes)))
    print('  duration   = {0}'.format(format_seconds(duration)))


if __name__ == '__main__':
    main()
