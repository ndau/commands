#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh
}

export WHT="\\33[22;37m"
export TEAL="\\33[22;36m"
export BLK="\\33[22;30m"
export RED="\\33[22;31m"
export GRN="\\33[22;32m"
export YEL="\\33[22;33m"
export BLU="\\33[22;34m"
export MAG="\\33[22;35m"
export CYN="\\33[22;36m"
export BRED="\\33[1;31m"
export BGRN="\\33[1;32m"
export BYEL="\\33[1;33m"
export BBLU="\\33[1;34m"
export BMAG="\\33[1;35m"
export BCYN="\\33[1;36m"
export BWHT="\\33[1;37m"
export BBLK="\\33[1;30m"
export BGBLK="\\33[40m"
export BGRED="\\33[41m"
export BGGRN="\\33[42m"
export BGYEL="\\33[43m"
export BGBLU="\\33[44m"
export BGMAG="\\33[45m"
export BGCYN="\\33[46m"
export BGWHT="\\33[47m"
export KILLCOLOR="\\33[0m"

status_one() {
    DIR=$NDEV_DIR/$1
    if [ -d "$DIR" ]; then
        cd "$DIR" || exit 1

        branch=$("$CMDBIN_DIR"/branch.sh)
        if [ "$branch" = "[ master ]" ]; then
            brcol=$GRN
        else
            brcol=$YEL
        fi

        changed=$(git status -s |wc -l)
        if [ "$changed" = 0 ]; then
            chgcol=$GRN
            msg="       unchanged"
        else
            chgcol=$YEL
            msg=$(printf "%2d files changed" "$changed")
        fi
        printf "%17s: $brcol %16s $KILLCOLOR $chgcol(%s)$KILLCOLOR\\n" "$1" "$branch" "$msg"
    else
        printf "%17s: %18s $RED(%16s)$KILLCOLOR\\n" "$1" " " "not present"
    fi
}

initialize
for f in {automation,chaincode,chaos,commands,integration-tests,metanode,ndau,ndaumath}; do
    status_one "$f"
done
