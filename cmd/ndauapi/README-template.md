# ndauapi

This tool provides an http interface to an ndau node.

# Design

This tool uses a [boneful](https://github.com/kentquirk/boneful) service, based on the [bone router](https://github.com/go-zoo/bone).

Configuration is provided with environment variables specifying the following

  * How much logging you want (error, warn, info, debug).
  * The protocol, host and port of the ndau node's rpc port. Required.
  * And the port to listen on.

Communication between this program and tendermint is firstly done with the tool pkg and indirectly with [Tendermint's RPC client](https://github.com/tendermint/tendermint/tree/master/rpc/client), which is based on JSON RPC.

Testing depends on a test net to be available and as such are not very pure unit tests. A TODO is to find a suitable mock.

# Getting started

```shell
./build.sh
NDAUAPI_NDAU_RPC_URL=http://127.0.0.1:31001 NDAUAPI_CHAOS_RPC_URL=http://127.0.0.1:31005 ./ndauapi
```

# Basic Usage

```shell
# get node status
curl localhost:3030/status
```

# Testing in VSCode

Please include this in your VSCode config to run individual tests. Replace the IP and port with your ndau node's IP and Tendermint RPC port.

```json
    "go.testEnvVars": {
        "NDAUAPI_NDAU_RPC_URL": "http://127.0.0.1:31001",
        "NDAUAPI_CHAOS_RPC_URL": "http://127.0.0.1:31005"
    },
```

# API

The following is automatically generated. Please do not edit the README.md file. Any changes above this section should go in (README-template.md).

> TODO: Please note that this documentation implementation is incomplete and is missing complete responses.

api_replacement_token
