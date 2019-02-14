#!/usr/bin/env python3

"""
Verify the ETL process against a running blockchain.

The mandate here is fairly simple:

- read the genesis file
- for each account, verify the following fields by comparing genesis to blockchain:
    - balance
    - last EAI date

Note: this tool assumes that the ndau tool is available and correctly configured.
"""

import csv
import functools
import json
import subprocess
import toml
import sys

from dateutil import parser as dtparser
from datetime import timezone
from pathlib import Path


def get_account_data(address):
    query = subprocess.run(
        ["./ndau", "account", "query", "-a", address],
        capture_output=True,
        text=True,
        check=True,
    )
    ad = json.loads(query.stdout)
    return ad


@functools.lru_cache(1)
def config(path="config.toml"):
    with open(path, "r") as f:
        return toml.load(f)


@functools.lru_cache(1)
def get_headers(conf_path):
    # headers appear two rows before the data begins
    line_no = config(conf_path)["first_row"] - 2

    with open(csv_path(conf_path), "r") as f:
        for i, line in enumerate(f):
            if i == line_no:
                return line.split(",")


@functools.lru_cache(1)
def csv_path(conf_path):
    return Path(conf_path).parent / config(conf_path)["path"]


def get_rows(conf_path):
    with open(csv_path(conf_path), "r") as f:
        reader = csv.DictReader(f, fieldnames=get_headers(conf_path))
        for i, row in enumerate(reader):
            if i >= config(conf_path)["first_row"] and len(row["address ID"]) > 0:
                yield row


def verify_row(row, verbose=False):
    addr = row["address ID"].strip()
    if verbose:
        print(f"{addr}... ", end="", flush=True)

    acct_data = get_account_data(addr)

    mismatch = []

    expect_balance = int(float(row["ndau amount in"]) * 100_000_000)
    actual_balance = acct_data["balance"]
    if expect_balance != actual_balance:
        mismatch.append(f"balance: want {expect_balance}, have {actual_balance}")

    expect_date = dtparser.parse(row["chain date"])
    if expect_date.tzinfo is None:
        expect_date = expect_date.replace(tzinfo=timezone.utc)
    actual_date = dtparser.isoparse(acct_data["lastEAIUpdate"])
    assert actual_date.tzinfo is not None
    if expect_date != actual_date:
        mismatch.append(
            f"acct date: want {expect_date.isoformat()}, have {actual_date.isoformat()}"
        )

    errs = "; ".join(mismatch)

    if verbose:
        if mismatch == "":
            print("OK")
        else:
            print(errs)

    return errs


def verify(conf_path, verbose=False):
    all_ok = True
    for row in get_rows(conf_path):
        err = verify_row(row, verbose)
        if err != "":
            all_ok = False
            if not verbose:
                print(err, file=sys.stderr)

    code = 0 if all_ok else 1
    sys.exit(code)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="verify post-etl pre-eai account state"
    )
    parser.add_argument("-v", "--verbose", action="store_true", help="say more")
    parser.add_argument(
        "-c",
        "--conf-path",
        default="config.toml",
        help="path to etl config.toml (default: ./config.toml)",
    )

    args = parser.parse_args()
    verify(args.conf_path, args.verbose)
