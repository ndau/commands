#!/bin/bash

# Stop building later steps if an earlier step fails.
set -e

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/ndau/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh
}

# Ensure we have no vendor links for tests active, as they will fail the build.
ensure_no_test_link() {
    REPO="$1"

    # Use -e not -L for extra robustness.  There shouldn't be anything named vendor there.
    if [ -e "$NDEV_DIR/$REPO/vendor" ]; then
        unlink_vendor_for_test "$REPO"
    fi
}

ensure_no_test_links() {
    ensure_no_test_link ndau
}

escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e's/\n/\\n/g'
}

build_ndau() {
    echo building ndau
    cd "$COMMANDS_DIR"

    ensure_no_test_links

    # Get the version info from git (we want the most recent tag that starts with v)
    # then use it to stamp the ndau executable as part of the build.
    VERSION=$(git describe --long --tags --match="v*")
    echo "  VERSION=$VERSION"
    VERSION_PKG="$NDEV_SUBDIR/ndau/pkg/version"
    go version
    export GO111MODULE=on
    go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./"$NDAU_CMD"
    go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./"$NDSH_CMD"
    go build -ldflags "-X $VERSION_PKG.version=$VERSION" ./"$NDAUNODE_CMD"
    go build ./"$NDAUAPI_CMD"

    # generate api documentation
    api_doc="$(escape_newlines "$(./ndauapi -docs)")"
    tmpl="$(escape_newlines "$(cat "$NDAUAPI_CMD/README-template.md")")"

    # generate new readme with api documentation
    readme="${tmpl/api_replacement_token/$api_doc}"
    echo -e "$readme" > "$NDAUAPI_CMD/README.md"
}

build_tools() {
    echo building tools
    cd "$COMMANDS_DIR"

    go build ./"$ETL_CMD"
    go build ./"$GENERATE_CMD"
    go build ./"$KEYTOOL_CMD"
    go build ./"$PROCMON_CMD"
    go build ./"$CLAIMER_CMD"
}

build_tm() {
    echo building tendermint
    cd "$TENDERMINT_DIR"
    # JSG move to make to satisfy new go dependency reqs in v0.32.5, we might need to go back
    # to "go build" in the future
#    echo "$TENDERMINT_DIR"
#    printenv
#    go get -u golang.org/x/sys # Update to support go 1.18 builds
#    go build ./"$TENDERMINT_CMD"
    go version
    GO111MODULE=on make build
}

build_noms() {
    echo building noms
    cd "$NOMS_DIR"
    GO111MODULE=on go build ./"$NOMS_CMD"
#    go build ./"$NOMS_CMD"
}

build_all() {
    initialize
    build_noms
    build_tm
    build_ndau
    build_tools
}

build_all
echo "done."
