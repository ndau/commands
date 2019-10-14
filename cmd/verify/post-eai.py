#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

"""
Automatic verification for accounts after the first CreditEAI tx

The main goal here is to consume the spreadsheet CSV and emit a new one,
containing all its current data plus some new columns:

- credited EAI
- currency seats in date order (ordinal)

Note: this tool assumes that the ndau tool is available and correctly configured.
"""

import csv
import sys

from collections import OrderedDict
from util import get_account_data, get_currency_seats, get_headers, get_rows


def append_issued_eai_col(rows_iter):
    """
    Iterator adaptor appending an "issued eai" column to each row of a rows iterator.

    This is a generator.
    """
    for row in rows_iter:
        addr = row["address ID"].strip()
        try:
            acct_data = get_account_data(addr)
        except Exception as e:
            print(e, file=sys.stderr)
            continue

        original_balance = int(float(row["ndau amount in"]) * 100_000_000)
        current_balance = acct_data["balance"]
        row["issued eai"] = current_balance - original_balance
        yield row


def append_currency_seat_ordinal_col(rows_iter):
    """
    Iterator adaptor appending a "currency seat ordinal" column to each row of
    a rows iterator.

    This is a generator.
    """
    rmap = OrderedDict()
    for row in rows_iter:
        rmap[row["address ID"]] = row

    try:
        for idx, seat in enumerate(get_currency_seats()):
            try:
                rmap[seat]["currency seat ordinal"] = idx
            except KeyError:
                # if the row map doesn't know about this currency seat, that's
                # ok; it must have arrived from some process other than ETL
                continue
    except Exception as e:
        print("failed to get currency seats:", e, file=sys.stderr)

    for row in rmap.values():
        yield row


def pipeline(iterator, *adaptors):
    "Apply a bunch of adaptors to an iterator."
    for adaptor in adaptors:
        iterator = adaptor(iterator)
    return iterator


def headers(conf_path):
    return get_headers(conf_path) + ["issued eai", "currency seat ordinal"]


def csv_out(rows_iter, conf_path):
    writer = csv.DictWriter(sys.stdout, fieldnames=headers(conf_path))

    writer.writeheader()
    writer.writerows(rows_iter)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="verify post-etl pre-eai account state"
    )
    parser.add_argument(
        "-c",
        "--conf-path",
        default="config.toml",
        help="path to etl config.toml (default: ./config.toml)",
    )

    args = parser.parse_args()

    rows = pipeline(
        get_rows(args.conf_path),
        append_issued_eai_col,
        append_currency_seat_ordinal_col,
    )
    csv_out(rows, args.conf_path)
