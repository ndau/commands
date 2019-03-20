SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

mkdir "$NODE_DATA_DIR"

cd "$BIN_DIR" || exit 1

echo Configuring intra-nodegroup port references for chaosnode and ndaunode...
./chaosnode --set-ndaunode "http://localhost:$TM_NDAU_RPC_PORT"
./ndaunode --set-chaosnode "http://localhost:$TM_CHAOS_RPC_PORT"

echo Importing genesis data into ndau conf...
# Build up a temporary genesis file to leverage ndaunode's updte-conf-from feature.
genesis_toml="$SCRIPT_DIR"/genesis.toml
echo '["'"$SVI_NAMESPACE"'"]' > "$genesis_toml"
echo '  ["'"$SVI_NAMESPACE"'".c3Zp]' >> "$genesis_toml"
echo '    svi_stub = true' >> "$genesis_toml"
./ndaunode -use-ndauhome -update-conf-from "$genesis_toml"
rm -f "$genesis_toml"

echo Removing ndau mock setting...
sed -i \
    -e "/UseMock/d" \
    "$NODE_DATA_DIR/ndau/config.toml"

echo Configuring tendermint...
#./tendermint init --home "$TM_CHAOS_DATA_DIR"
#./tendermint init --home "$TM_NDAU_DATA_DIR"
#sed -i -E \
#    -e 's/^(create_empty_blocks = .*)/# \1/' \
#    -e 's/^(create_empty_blocks_interval =) (.*)/\1 "300s"/' \
#    -e 's/^(addr_book_strict =) (.*)/\1 false/' \
#    -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
#    "$TM_CHAOS_DATA_DIR/config/config.toml" \
#    "$TM_NDAU_DATA_DIR/config/config.toml"

echo Configuration complete
