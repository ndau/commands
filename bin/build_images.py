#!/usr/bin/env python3

import os
import re
import shlex
import subprocess
import sys
from functools import lru_cache

import requests

VERSION = re.compile(r"^v[.\d]+(-\w[-.\w]*)?$")


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


def current_branch() -> str:
    return run("git symbolic-ref --short HEAD", stderr=subprocess.DEVNULL, check=False)


@lru_cache(1)
def is_version_tag(branch) -> bool:
    return VERSION.match(branch) is not None


@lru_cache(1)
def commands_sha(branch="HEAD") -> str:
    if branch != "HEAD":
        run("git fetch origin", timeout=30)
        if not is_version_tag(branch):
            branch = f"origin/{branch}"
    return run(f"git rev-parse --short {branch}")


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


def prepare_commands(branch) -> None:
    with open(rooted("docker", "build_commands", "commands_sha"), "wt") as fp:
        print(commands_sha(), file=fp)
    for pkgfile in ("Gopkg.toml", "Gopkg.lock"):
        with open(rooted("docker", "build_commands", pkgfile), "wt") as fp:
            fp.write(run(f"git show {branch}:{pkgfile}"))


class BuildError(Exception):
    def __init__(self, image):
        super(BuildError, self).__init__(f"failed to build {image}")
        self.image = image


class PushError(Exception):
    def __init__(self, tag):
        super(BuildError, self).__init__(f"failed to push {tag}")
        self.tag = tag


def build(
    image: str, branch: str, env: dict = {}, public: bool = False, push: bool = False
) -> None:
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
    ptags = []
    if public:
        ptags = [
            f"oneirondev/{image}:{commands_sha(branch)}",
            f"oneirondev/{image}:latest",
        ]
        if is_version_tag(branch):
            ptags.append(f"oneirondev/{image}:{branch}")
        for ptag in ptags:
            cmd.append(f"--tag={ptag}")

    try:
        run(cmd, timeout=None)
    except subprocess.CalledProcessError as e:
        print(e.stdout, file=sys.stderr)
        raise BuildError(image)

    if push:
        for tag in ptags:
            cmd = ["docker", "push", tag]
            try:
                run(cmd, timeout=None)
            except subprocess.CalledProcessError as e:
                print(e.stdout, file=sys.stderr)
                raise PushError(tag)


def main(branch: str, run_unit_tests: bool, push: bool) -> None:
    if run("git status --porcelain") != "" and branch == current_branch():
        print(
            "WARNING: docker images contain only committed and pushed work "
            f"({commands_sha(branch)})"
        )

    def sbuild(*args, **kwargs):
        "build, handling build errors"
        args = list(args)
        args.insert(1, branch)
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
    prepare_commands(branch)
    env = {"COMMANDS_BRANCH": branch}
    if run_unit_tests:
        env["RUN_UNIT_TESTS"] = "1"
    sbuild("build_commands", env=env)

    # prepare and build the ndaunode and integration tests public images
    sbuild("ndauimage", public=True, push=push)
    sbuild("integration_tests")

    # prepare and build a tools image
    sbuild("tools", public=True, push=push)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="build ndau docker images")
    parser.add_argument(
        "branch",
        default=current_branch(),
        nargs="?",
        help=("build this branch/tag of the commands repo. default: current branch"),
    )
    parser.add_argument(
        "--run-unit-tests",
        action="store_true",
        help="run unit tests after building the commands packages",
    )
    parser.add_argument(
        "--push",
        action="store_true",
        help=(
            "push public images to docker hub after generation. "
            "note: must have stored credentials with `docker login`"
        ),
    )

    args = parser.parse_args()
    main(args.branch, args.run_unit_tests, args.push)
