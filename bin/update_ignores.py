#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import pathlib
import subprocess


def findroot():
    completed = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        check=True,
        text=True,
        capture_output=True,
        timeout=1,
    )
    return pathlib.Path(completed.stdout.strip())


ROOT = findroot()


def gitignore_lines():
    with (ROOT / ".gitignore").open() as gitignore:
        return set(line.rstrip() for line in gitignore)


def commands():
    yield from (c.parts[-1] for c in sorted((ROOT / "cmd").iterdir()) if c.is_dir())


def unignore(cmd):
    return f"!**/{cmd}/"


def update():
    gil = gitignore_lines()

    with (ROOT / ".gitignore").open("a") as gitignore:
        for cmd in commands():
            if cmd not in gil:
                print(cmd, file=gitignore)
            u = unignore(cmd)
            if u not in gil:
                print(u, file=gitignore)


if __name__ == "__main__":
    update()
