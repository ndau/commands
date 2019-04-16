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
            brcol=$YEL
        fi

        origincommit=$(git log --pretty="%h" -n 1 origin/"$branch" 2> /dev/null)
        localcommit=$(git log --pretty="%h" -n 1)

        changed=$(git status -s |wc -l|tr -Cd "[:digit:]")
        if [ "$changed" = "0" ]; then
            if [ "$localcommit" = "$origincommit" ]; then
                chgcol=$GRN
                msg="unchanged"
            else
                chgcol=$BYEL
                msg="out of date"
            fi
        else
            onlynew=$(git status --porcelain |grep -v "??" |wc -l|tr -Cd "[:digit:]")
            if [ "$onlynew" = "0" ]; then
                chgcol=$BLU
            else
                chgcol=$BYEL
            fi
            details=$(git status --porcelain |cut -c 1,2 |sort |uniq -c |tr -s "\n" "," |tr -d " "|sed "s/??/ new/"|sed s/,$//)
            msg=$(printf "%d changed - %s" "$changed" "$details")
        fi

        if [ "$brcol" = "$chgcol" ]; then
            dircol=$brcol
        else
            dircol=$BWHT
        fi
        printf "$dircol%24s: $brcol %19s $KILLCOLOR $chgcol(%s)$KILLCOLOR\\n" "$1" "$branch" "$msg"
    else
        printf "%24s: %18s $RED(%19s)$KILLCOLOR\\n" "$1" " " "not present"
    fi
}

initialize

if [ "$1" = "--all" ]; then
    for f in "$NDEV_DIR"/*; do
        if [ -d $f ]; then
            if [ -d "$f/.git" ]; then
                status_one "$(basename "$f")"
            fi
        fi
    done
else
    for f in {automation,chaincode,commands,generator,integration-tests,metanode,ndau,ndaumath,o11y}; do
        status_one "$f"
    done
fi
