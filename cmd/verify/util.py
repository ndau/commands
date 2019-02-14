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
