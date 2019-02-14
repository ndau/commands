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
import requests
import sys
import toml

from dateutil import parser as dtparser
from datetime import timezone
from pathlib import Path


def requests_retry_session(
    retries=3, backoff_factor=0.3, status_forcelist=(500, 502, 504), session=None
):
    from requests.adapters import HTTPAdapter
    from requests.packages.urllib3.util.retry import Retry

    session = session or requests.Session()
    retry = Retry(
        total=retries,
        read=retries,
        connect=retries,
        backoff_factor=backoff_factor,
        status_forcelist=status_forcelist,
    )
    adapter = HTTPAdapter(max_retries=retry)
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    return session


def get_account_data(host, address):
    response = requests_retry_session().get(f"{host}/account/account/{address}")
    response.raise_for_status()
    return response.json()


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


def verify_row(host, row, verbose=False):
    addr = row["address ID"].strip()
    if verbose:
        print(f"{addr}... ", end="", flush=True)

    acct_data = get_account_data(host, addr)

    mismatch = []

    expect_balance = int(float(row["ndau amount in"]) * 100_000_000)
    actual_balance = acct_data[addr]["balance"]

    if expect_balance != actual_balance:
        mismatch.append(f"balance: want {expect_balance}, have {actual_balance}")

    expect_date = dtparser.parse(row["chain date"])
    if expect_date.tzinfo is None:
        expect_date = expect_date.replace(tzinfo=timezone.utc)
    actual_date = dtparser.isoparse(acct_data[addr]["lastEAIUpdate"])
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


def verify(host, conf_path, verbose=False):
    all_ok = True
    for row in get_rows(conf_path):
        try:
            err = verify_row(host, row, verbose)
        except requests.exceptions.HTTPError as e:
            err = str(e)

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
    parser.add_argument(
        "--host",
        action="store",
        default="http://localhost:3030",
        dest="host",
        help="specify an arbitrary host location",
    )
    parser.add_argument(
        "--main",
        action="store_const",
        dest="host",
        const="https://node-0.main.ndau.tech",
        help="use mainnet as host",
    )
    parser.add_argument(
        "--test",
        action="store_const",
        dest="host",
        const="https://testnet-0.api.ndau.tech",
        help="use testnet as host",
    )
    parser.add_argument(
        "--dev",
        action="store_const",
        dest="host",
        const="https://devnet-0.api.ndau.tech",
        help="use devnet as host",
    )
    parser.add_argument(
        "--local",
        action="store_const",
        dest="host",
        const="http://localhost:3030",
        help="use localhost:3030 as host",
    )

    args = parser.parse_args()
    verify(args.host, args.conf_path, args.verbose)
