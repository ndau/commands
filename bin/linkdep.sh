#!/bin/bash

# Create symlinks from the commands vendor directories to the given dependency repo.
# Useful if you want to make changes to a dependency project and test locally.
# Run this after doing a `dep ensure` from the commands repo.
#
# Usage example:
#   ./linkdep.sh metanode
#
# Assumes you've already cloned the dependency repo next to chaos and ndau repos.
# This script can be run from anywhere.
#
# Use `dep ensure` again to undo what this script did to the vendor subdirectories.

DEP="$1"
if [ -z "$DEP" ]; then
    echo Usage examples:
    echo "  ./linkdep.sh metanode"
    echo "  ./linkdep.sh chaos"
    echo "  ./linkdep.sh ndau"
    echo "  ./linkdep.sh all"
    exit 1
fi

GO_DIR=$(go env GOPATH)
if [[ "$GO_DIR" == *":"* ]]; then
    echo Multiple Go paths not supported
    exit 1
fi

NDEV_SUBDIR=github.com/oneiro-ndev
SOURCE_DIR="$GO_DIR/src/$NDEV_SUBDIR"

link_dep() {
    REPO="$1"

    DEP_VENDOR_DIR="$SOURCE_DIR/commands/vendor/$NDEV_SUBDIR/$REPO"
    DEP_SOURCE_DIR="$SOURCE_DIR/$REPO"

    if [ ! -d "$DEP_SOURCE_DIR" ]; then
        echo Must clone "$REPO" into "$DEP_SOURCE_DIR" first
        exit 1
    fi

    rm -rf "$DEP_VENDOR_DIR"
    ln -s "$DEP_SOURCE_DIR" "$DEP_VENDOR_DIR"
}

if [ "$DEP" == "all" ]; then
    link_dep metanode
    link_dep chaos
    link_dep ndau
else
    link_dep "$DEP"
fi
