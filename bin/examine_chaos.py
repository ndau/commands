#!/usr/bin/env python3

import subprocess
import shlex
from pathlib import Path
from base64 import standard_b64decode as b64d
from functools import lru_cache


def rc(cmd):
    return subprocess.run(
        shlex.split(cmd), capture_output=True, text=True, check=True, timeout=10
    )


@lru_cache(1)
def root():
    out = rc("git rev-parse --show-toplevel")
    return out.stdout.strip()


def chaos(cmd):
    chaos_bin = Path(root()) / "chaos"
    return rc(f"{chaos_bin} {cmd}")


def get_namespaces():
    out = chaos("get-ns")
    return out.stdout.splitlines()


def dump_ns(ns):
    output = chaos(f"dump --ns={ns}")
    out = {}
    for line in output.stdout.splitlines():
        k, v = line.split()
        out[b64d(k)] = b64d(v)
    return out


def collect_all():
    return {ns: dump_ns(ns) for ns in get_namespaces()}


def stringify(v):
    "recursively replace byte sequences with safe strings"
    if isinstance(v, str):
        return v
    elif isinstance(v, bytes):
        return v.decode(errors="backslashreplace")
    elif isinstance(v, list):
        return [stringify(vv) for vv in v]
    elif isinstance(v, tuple):
        return tuple([stringify(vv) for vv in v])
    elif isinstance(v, set):
        return {stringify(vv) for vv in v}
    elif isinstance(v, dict):
        return {stringify(k): stringify(vv) for k, vv in v.items()}
    else:
        raise Exception(f"can't stringify {type(v)}")


if __name__ == "__main__":
    import json

    print(json.dumps(stringify(collect_all())))
