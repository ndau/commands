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
        if [ "$branch" = "master" ]; then
            brcol=$GRN
        else
            # non-master branches are pointed out
            brcol=$YEL
        fi

        # get the commit hash for this branch and the equivalent at the origin (if it exists)
        origincommit=$(git log --pretty="%h" -n 1 origin/"$branch" 2> /dev/null)
        localcommit=$(git log --pretty="%h" -n 1)

        # count the number of files changed and strip any non-numerics
        changed=$(git status -s |wc -l|tr -Cd "[:digit:]")
        if [ "$changed" = "0" ]; then
            if [ "$localcommit" = "$origincommit" ]; then
                # no changes and up to date
                chgcol=$GRN
                msg="unchanged"
            else
                # no local changes but it's not up to date with origin
                chgcol=$BYEL
                msg="out of date"
            fi
        else
            # how many of the changes were something other than new files?
            nonnew=$(git status --porcelain |grep -v "??" |wc -l|tr -Cd "[:digit:]")
            # get details of modified, deleted, new, etc.
            details=$(git status --porcelain |cut -c 1,2 |sort |uniq -c |tr -s "\n" "," |tr -d " "|sed "s/??/ new/"|sed s/,$//)
            # if there were only new files, then we're not as concerned
            if [ "$nonnew" = "0" ]; then
                chgcol=$BLU
                msg=$(printf "%s" "$details")
            else
                # but other changes get a higher level of warning
                chgcol=$BYEL
                msg=$(printf "%d changed - %s" "$changed" "$details")
            fi
        fi

        if [ "$brcol" = "$chgcol" ]; then
            # if these two are both green, make the folder name green too
            dircol=$brcol
        else
            # otherwise make the folder more obvious
            dircol=$BWHT
        fi
        printf "$dircol%24s: $brcol %19s $KILLCOLOR $chgcol(%s)$KILLCOLOR\\n" "$1" "$branch" "$msg"
    else
        # an expected folder was missing
        printf "%24s: %18s $RED(%19s)$KILLCOLOR\\n" "$1" " " "not present"
    fi
}

initialize

# if --all is specified then do all the git folders underneath the oneiro repository
if [ "$1" = "--all" ]; then
    for f in "$NDEV_DIR"/*; do
        if [ -d "$f" ]; then
            if [ -d "$f/.git" ]; then
                status_one "$(basename "$f")"
            fi
        fi
    done
else
    # this is the list of expected folders
    for f in {automation,chaincode,commands,generator,integration-tests,json2msgp,metanode,ndau,ndaumath,o11y,recovery,system_vars,writers}; do
        status_one "$f"
    done
fi
