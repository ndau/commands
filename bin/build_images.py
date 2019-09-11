#!/usr/bin/env python3

import os
import shlex
import shutil
import subprocess
import sys
from functools import lru_cache

import requests


def run(cmd, timeout=1, check=True, stderr=subprocess.STDOUT, env={}) -> str:
    """
    Run a shell command.

    Wraps subprocess.run to set better defaults.
    """
    if isinstance(cmd, str):
        cmd = shlex.split(cmd)
    elif isinstance(cmd, list):
        pass
    else:
        raise ValueError("invalid type: cmd must be str or list")

    result = subprocess.run(
        cmd,
        timeout=timeout,
        check=check,
        stdout=subprocess.PIPE,
        stderr=stderr,
        text=True,
    )
    return result.stdout.strip()


@lru_cache(1)
def commands_sha() -> str:
    return run("git rev-parse --short HEAD")


def rooted(*components) -> str:
    "return path based on the root of the commands dir"
    return os.path.join(os.path.dirname(__file__), "..", *components)


def prepare_noms() -> None:
    resp = requests.get(
        "https://api.github.com/repos/attic-labs/noms/git/refs/heads/master"
    )
    resp.raise_for_status()
    with open(rooted("docker", "noms", "noms_sha"), "wt") as fp:
        print(resp.json()["object"]["sha"], file=fp)


def prepare_commands() -> None:
    with open(rooted("docker", "build_commands", "commands_sha"), "wt") as fp:
        print(commands_sha(), file=fp)


class BuildError(Exception):
    def __init__(self, image):
        super(BuildError, self).__init__(f"failed to build {image}")
        self.image = image


def build(image: str, env: dict = {}) -> None:
    """
    Build one of our docker images
    """
    print(f"building {image}...")
    if run(f"docker container ls -a -q -f ancestor={image}") != "":
        print(
            f"  WARNING: containers exist based on an old {image}; "
            "they should be removed"
        )

    cmd = ["docker", "build"]
    for k, v in env.items():
        cmd.append("--build-arg")
        cmd.append(f"{k}={v}")
    cmd.append(rooted("docker", image))
    cmd.append(f"--tag={image}:{commands_sha()}")
    cmd.append(f"--tag={image}:latest")

    try:
        run(cmd, timeout=None)
    except subprocess.CalledProcessError as e:
        print(e.stdout, file=sys.stderr)
        raise BuildError(image)


def main(branch: str, run_unit_tests: bool) -> None:
    if run("git status --porcelain") != "":
        print("WARN: uncommitted changes")
        print(f"docker image contains only committed work ({commands_sha()})")

    def sbuild(*args, **kwargs):
        "build, handling build errors"
        try:
            build(*args, **kwargs)
        except BuildError as e:
            print(e, file=sys.stderr)
            sys.exit(1)

    # prepare and build the base go build image
    with open(rooted("machine_user_key"), "rt") as fp:
        muk = fp.read()
    sbuild("go_build", env={"SSH_PRIVATE_KEY": muk})

    # prepare and build noms
    prepare_noms()
    sbuild("noms")

    # prepare and build the commands packages
    shutil.copy2(rooted("Gopkg.toml"), rooted("docker", "build_commands", "Gopkg.toml"))
    shutil.copy2(rooted("Gopkg.lock"), rooted("docker", "build_commands", "Gopkg.lock"))
    env = {"COMMANDS_BRANCH": branch}
    if run_unit_tests:
        env["RUN_UNIT_TESTS"] = "1"
    sbuild("build_commands", env=env)

    # prepare and build the ndaunode and integration tests public images
    sbuild("ndaunode")
    sbuild("integration_tests")


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="build ndau docker images")
    parser.add_argument(
        "--branch",
        default=run(
            "git symbolic-ref --short HEAD", stderr=subprocess.DEVNULL, check=False
        ),
        help=("build this branch/tag of the commands repo. default: current branch"),
    )
    parser.add_argument(
        "--run-unit-tests",
        action="store_true",
        help="run unit tests after building the commands packages",
    )

    args = parser.parse_args()
    main(args.branch, args.run_unit_tests)
