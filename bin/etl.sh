#!/bin/bash

echo ETL for ndau noms
# cd to genesis dir because etl command looks for ./config.toml there
echo setting CWD to $GO_DIR/src/github.com/ndau/genesis
pushd $NDEV_DIR/genesis/$CHAIN_ID-genesis/etl || exit 1
# if spreadsheet file doesn't exist, die
# if [ ! -e "$SPREADSHEET_FILE" ]; then
#    echo spreadsheet file: \"$SPREADSHEET_FILE\" does not exist, set variable \$SPREADSHEET_FILE to spreadsheet file name, exiting
#    exit 1
# fi
# process spreadsheet, merge user id and address data from dashboard MongoDB
# bin/process_csv.py -i "$SPREADSHEET_FILE"
# run etl, input is "output.csv" created in last step, data is written to ndau noms dir specified in ./config.toml file
$COMMANDS_DIR/etl
popd || exit 1
