# ndau API

> This API reference is automatically generated: do not edit it. Make changes in `README-template.md`.

The ndau API provides an http interface to an ndau node. It is used to prevalidate and submit transactions, and to retrieve information about the ndau blockchain's system variables, price information, transactions, blocks, and accounts.

api_replacement_token

## Design

This tool uses a [boneful](https://github.com/kentquirk/boneful) service, based on the [bone router](https://github.com/go-zoo/bone).

Configuration is provided with environment variables specifying the following

  * How much logging you want (error, warn, info, debug).
  * The protocol, host and port of the ndau node's rpc port. Required.
  * And the port to listen on.

Communication between this program and tendermint is firstly done with the tool pkg and indirectly with [Tendermint's RPC client](https://github.com/tendermint/tendermint/tree/master/rpc/client), which is based on JSON RPC.

Testing depends on a test net to be available and as such are not very pure unit tests.

## Getting started

```shell
./build.sh
NDAUAPI_NDAU_RPC_URL=http://127.0.0.1:31001 ./ndauapi
```

## Basic Usage

```shell
# get node status
curl localhost:3030/status
```

## Testing in VSCode

Please include this in your VSCode config to run individual tests. Replace the IP and port with your ndau node's IP and Tendermint RPC port.

```json
    "go.testEnvVars": {
        "NDAUAPI_NDAU_RPC_URL": "http://127.0.0.1:31001"
    },
```
