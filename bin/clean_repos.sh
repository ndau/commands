#!/usr/bin/env bash

set -e

cd "$(go env GOPATH)/src/github.com/oneiro-ndev/"

errs=$(mktemp)

for repo in ./*; do
    if [ -d "$repo/.git" ]; then
        (
            echo
            echo "organizing $repo..."
            cd "$repo"
            # get the current master branch
            if git checkout master; then
                if ! git pull; then
                    echo "$repo: could not pull most recent" >> "$errs"
                fi
            else
                echo "$repo: could not checkout master" >> "$errs"
            fi
            # clean up dead tracking branches
            git fetch --prune
            for gone_branch in $(
                git branch -v |\
                grep '\[gone\]' |\
                tr -s ' ' |\
                cut -d' ' -f2
            ); do
                git branch -d "$gone_branch"
            done
        )
    fi
done

rv=0
if [ -s "$errs" ]; then
    rv=1
    echo
    echo "Errors encountered:"
    sed -e 's|^(./)?| |' "$errs"
fi
rm -f "$errs"

exit "$rv"
