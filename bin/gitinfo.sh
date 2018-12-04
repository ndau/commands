#!/bin/bash

WHT="\33[22;37m"
TEAL="\33[22;36m"
BLK="\33[22;30m"
RED="\33[22;31m"
GRN="\33[22;32m"
YEL="\33[22;33m"
BLU="\33[22;34m"
MAG="\33[22;35m"
CYN="\33[22;36m"
BRED="\33[1;31m"
BGRN="\33[1;32m"
BYEL="\33[1;33m"
BBLU="\33[1;34m"
BMAG="\33[1;35m"
BCYN="\33[1;36m"
BWHT="\33[1;37m"
BBLK="\33[1;30m"
BGBLK="\33[40m"
BGRED="\33[41m"
BGGRN="\33[42m"
BGYEL="\33[43m"
BGBLU="\33[44m"
BGMAG="\33[45m"
BGCYN="\33[46m"
BGWHT="\33[47m"
KILLCOLOR="\33[0m"

status_one() {
    DIR=$NDEV_DIR/$1
    if [ -d $DIR ]; then
        cd $DIR

        branch=`$SETUP_DIR/branch.sh`
        if [ "$branch" = "[ master ]" ]; then
            brcol=$GRN
        else
            brcol=$YEL
        fi

        changed=$(git status -s |wc -l)
        if [ $changed = 0 ]; then
            chgcol=$GRN
            msg="       unchanged"
        else
            chgcol=$YEL
            msg=$(printf "%2d files changed" $changed)
        fi
        printf "%15s: $brcol %16s $KILLCOLOR $chgcol(%s)$KILLCOLOR\n" "$1" "$branch" "$msg"
    else
        printf "%15s: %18s $RED(%16s)$KILLCOLOR\n" $1 " " "not present"
    fi
}

SETUP_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
source $SETUP_DIR/env.sh
for f in {automation,chaincode,chaos,chaos_genesis,metanode,ndau,ndaumath,noms}; do
    status_one $f
done
