#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import csv
import functools
import json
import subprocess
import toml

from pathlib import Path


def get_account_data(address):
    query = subprocess.run(
        ["./ndau", "account", "query", "-a", address],
        capture_output=True,
        text=True,
        check=True,
        timeout=3,  # seconds
    )
    ad = json.loads(query.stdout)
    return ad


def get_currency_seats():
    query = subprocess.run(
        ["./ndau", "currency-seats"], capture_output=True, text=True, timeout=10
    )
    return [line for line in query.stdout.splitlines() if line != ""]


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
                return line.strip().split(",")


@functools.lru_cache(1)
def csv_path(conf_path):
    return Path(conf_path).parent / config(conf_path)["path"]


def get_rows(conf_path):
    with open(csv_path(conf_path), "r") as f:
        reader = csv.DictReader(f, fieldnames=get_headers(conf_path))
        for i, row in enumerate(reader):
            if i >= config(conf_path)["first_row"] and len(row["address ID"]) > 0:
                yield row
