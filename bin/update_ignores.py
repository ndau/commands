#!/usr/bin/env python3

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
