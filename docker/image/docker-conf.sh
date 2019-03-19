SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

GENESIS_TOML="$SCRIPT_DIR/genesis.toml"

mkdir -p "$NODE_DATA_DIR"

cd "$BIN_DIR" || exit 1

echo Configuring tendermint...
./tendermint init --home "$TENDERMINT_CHAOS_DATA_DIR"
./tendermint init --home "$TENDERMINT_NDAU_DATA_DIR"
sed -i -E \
    -e 's/^(create_empty_blocks = .*)/# \1/' \
    -e 's/^(create_empty_blocks_interval =) (.*)/\1 "300s"/' \
    -e 's/^(addr_book_strict =) (.*)/\1 false/' \
    -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
    "$TENDERMINT_CHAOS_DATA_DIR/config/config.toml" \
    "$TENDERMINT_NDAU_DATA_DIR/config/config.toml"

echo Configuring intra-nodegroup port references for chaosnode and ndaunode...
./chaosnode --set-ndaunode "http://localhost:$TM_NDAU_RPC_PORT"
./ndaunode --set-chaosnode "http://localhost:$TM_CHAOS_RPC_PORT"

echo Importing genesis data into chaos noms...
./genesis -g "$GENESIS_TOML" -n "$NOMS_CHAOS_DATA_DIR"
rm -f genesis

echo Importing genesis data into ndau conf...
./ndaunode -use-ndauhome -update-conf-from "$GENESIS_TOML"
mkdir "$NOMS_NDAU_DATA_DIR"

echo Removing ndau mock setting...
sed -i \
    -e "/UseMock/d" \
    "$NODE_DATA_DIR/ndau/config.toml"

echo Configuration step complete
