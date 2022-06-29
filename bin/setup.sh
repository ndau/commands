#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/ndau/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Process command line arguments.
node_count="$1"
chain_id="$2"
SNAPSHOT="$3"
if [ -z "$node_count" ]; then
    node_count=1
    echo "node_count not set; defaulting to $node_count"
else
    if [[ ! "$node_count" =~ ^[0-9]+$ ]]; then
        echo Node count must be a positive integer
        exit 1
    fi
    if [ "$node_count" -lt 1 ] || [ "$node_count" -gt "$MAX_NODE_COUNT" ]; then
        echo Node count must be in [1, "$MAX_NODE_COUNT"]
        exit 1
    fi
fi
if [ -z "$chain_id" ]; then
    chain_id=localnet
    echo "chain_id not set; defaulting to $chain_id"
fi

# Users may want us to generate the genesis files, or they may want to use their own.
# Checking this early on gives the user the chance to fix their mistake if they didn't want them
# generated.  It'll only ask once, even on subsequent setup.sh commands.
# Only check for for the system vars toml since the system accounts toml is optional.
if [[ ! -f "$SYSTEM_VARS_TOML" && -z "$SNAPSHOT" ]]; then
    echo "Cannot find genesis file: $SYSTEM_VARS_TOML"

    printf "Generate new? [y|n]: "
    read GENERATE
    if [ "$GENERATE" != "y" ]; then
        echo "Cannot set up a localnet without genesis files"
        echo "See instructions in ../README.md if you would like to use specific genesis files"
        exit 1
    fi

    # At this point, conf.sh will see that the genesis file (system vars toml) is missing and
    # will generate it as well as generating a fresh system accounts toml.
fi

# Initialize global config for the localnet we're setting up.
echo SETUP: Initializing a "$node_count"-node localnet...
# Ensure a fresh data directory.
rm -rf "$ROOT_DATA_DIR"
mkdir -p "$ROOT_DATA_DIR"
echo "$node_count" > "$NODE_COUNT_FILE"
echo "$chain_id" > "$CHAIN_ID_FILE"

# Get the correct version of noms source.
# mkdir -p "$ATTICLABS_DIR"
cd "$NDEV_DIR"
if [ -d "noms" ]; then
    cd noms
    ORIGIN_URL=$(git config --get remote.origin.url)
    if [ "$ORIGIN_URL" == "$NOMS_REPO" ]; then
        echo SETUP: Updating noms...
        git checkout master
        git pull origin master
    else
        echo SETUP: Replacing unsupported noms repo...
        cd ..
        rm -rf noms
        git clone "$NOMS_REPO"
    fi
else
    echo SETUP: Cloning noms...
    git clone "$NOMS_REPO"
fi
# Do a "go get" after cloning noms to match /docker/image/Dockerfile behavior.
echo SETUP: Getting noms...
go get -u "$NDEV_DIR"/noms/...

run_dep_ensure() {
    # These vendor directories sometimes cause dep ensure to fail, remove them first.
    rm -rf vendor
    rm -rf .vendor-new
#    "$GO_DIR"/bin/dep ensure
}

# Get the correct version of tendermint source.
echo SETUP: Getting dep...
# go get -u github.com/golang/dep/...
mkdir -p "$TM_DIR"
cd "$TM_DIR"
if [ -d "tendermint" ]; then
    echo SETUP: Updating tendermint...
    cd tendermint
    # Simulate same state as the else case for consistency and to prepare for version checkout.
    git checkout -- .
    git checkout master
    git pull origin master
else
    echo SETUP: Cloning tendermint...
    git clone "$TENDERMINT_REPO"
    cd tendermint
fi
echo SETUP: Checking out tendermint "$TENDERMINT_VER"...
git fetch --prune
git checkout "$TENDERMINT_VER"
# TODO: dep is no longer supported by tendermint; replace this.
#echo SETUP: Ensuring dependencies for tendermint...
#run_dep_ensure

# Get the ndev repos.
update_repo() {
    repo="$1"
    cd "$NDEV_DIR"
    if [ -d "$repo" ]; then
        echo SETUP: Updating "$repo"...
        cd "$repo"
        branch=$("$CMDBIN_DIR"/branch.sh)
        exists=$(git ls-remote --heads https://github.com/ndau/"$repo".git "$branch")
        if [ -z "$exists" ]; then
            # This just means you have a local branch you haven't pushed yet, and that's fine.
            echo "Branch $branch does not exist on remote"
        else
            git pull origin "$branch"
        fi
    else
        echo SETUP: Cloning "$repo"...
        git clone https://github.com/ndau/"$repo".git
    fi
}

mkdir -p "$NDEV_DIR"
update_repo commands
# We need the ndau repo only for running its unit tests from test.sh.
update_repo ndau

cd "$NDEV_DIR"/commands
echo SETUP: Ensuring dependencies for commands...
# run_dep_ensure

# Build everything.
echo SETUP: Building...
"$CMDBIN_DIR"/build.sh

# Test everything.
echo SETUP: Testing...
"$CMDBIN_DIR"/test.sh

# Configure everything.
echo SETUP: Configuring...
if [ -z "$SNAPSHOT" ]; then
    "$CMDBIN_DIR"/conf.sh --needs-update
else
    "$CMDBIN_DIR"/conf.sh --snapshot $SNAPSHOT
fi 

echo SETUP: Setup complete
