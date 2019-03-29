#!/bin/bash

# Create symlinks from the commands vendor directories to the given dependency repo.
# Useful if you want to make changes to a dependency project and build or test locally.
# NOTE: Running `dep ensure` from the commands repo may undo any links previously made.

CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Repo name.
REPO=""

# Linking or unlinking.
IS_LINKING=1

# For building or testing.
FOR_BUILDING=1

ARGS=("$@")
for arg in "${ARGS[@]}"; do
    # It's possible to specify conflicting arguments.  The later one "wins".
    if [ "$arg" = "-l" ] || [ "$arg" = "--link" ]; then
        IS_LINKING=1
    elif [ "$arg" = "-u" ] || [ "$arg" = "--unlink" ]; then
        IS_LINKING=0
    elif [ "$arg" = "-b" ] || [ "$arg" = "--build" ]; then
        FOR_BUILDING=1
    elif [ "$arg" = "-t" ] || [ "$arg" = "--test" ]; then
        FOR_BUILDING=0
    else
        REPO="$arg"
    fi
done

if [ -z "$REPO" ]; then
    echo linkdep: Link or unlink vendor directories of a repo, for build or test purposes
    echo Usage:
    echo "  ./linkdep.sh {metanode|ndau|genesis|all} [-l|--link] [-u|--unlink] [-b|--build] [-t|--test]"
    current="$(find vendor/github.com/oneiro-ndev -type l -depth 1 |sed 's:vendor/github.com/oneiro-ndev/:  :')"
    if [[ -n "$current" ]]; then
        echo Currently linked:
        echo "$current"
    fi
    exit 1
fi

if [ "$REPO" = "all" ]; then
    repos=(metanode ndau genesis)
else
    repos=("$REPO")
fi

for repo in "${repos[@]}"; do
    if [ "$FOR_BUILDING" != 0 ]; then
        if [ "$IS_LINKING" != 0 ]; then
            link_vendor_for_build "$repo"
        else
            unlink_vendor_for_build "$repo"
        fi
    else
        if [ "$IS_LINKING" != 0 ]; then
            link_vendor_for_test "$repo"
        else
            unlink_vendor_for_test "$repo"
        fi
    fi
done
