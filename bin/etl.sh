#!/bin/bash

echo ETL for ndau noms
# don't run ETL if we've updated this node already
echo "$NEEDS_UPDATE_FLAG_FILE-$1"
if [ ! -e "$NEEDS_UPDATE_FLAG_FILE-$1" ]; then
    echo ETL already run on node, unset \$RUN_ETL var to continue, exiting
    exit 1
fi
# don't run ETL if there is more than 1 node
if [ "$NODE_COUNT" -gt 1 ]; then
    echo ETL only runs in single node, please run bin/setup.sh 1, exiting
    exit 1
fi
# cd to genesis dir because etl command looks for ./config.toml there
echo setting CWD to $GO_DIR/src/github.com/oneiro-ndev/genesis
pushd $NDEV_DIR/genesis || exit 1
# if spreadsheet file doesn't exist, die
if [ ! -e "$SPREADSHEET_FILE" ]; then
    echo spreadsheet file: \"$SPREADSHEET_FILE\" does not exist, set variable \$SPREADSHEET_FILE to spreadsheet file name, exiting
    exit 1
fi
# process spreadsheet, merge user id and address data from dashboard MongoDB
bin/process_csv.py -i "$SPREADSHEET_FILE"
# run etl, input is "output.csv" created in last step, data is written to ndau noms dir specified in ./config.toml file
$COMMANDS_DIR/etl
popd || exit 1
