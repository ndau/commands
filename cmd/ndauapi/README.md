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


---
# `/`

This service provides the API for Tendermint and Chaos/Order/ndau blockchain data.

It is organized into several sections:

* /account returns data about specific accounts
* /block returns information about blocks on the blockchain
* /chaos returns information from the chaos chain
* /node provides information about node operations
* /order returns information from the order chain
* /transaction allows querying individual transactions on the blockchain
* /tx provides tools to build and submit transactions

Each of these, in turn, has several endpoints within it.




* [AccountByID](#accountbyid)

* [AccountsFromList](#accountsfromlist)

* [AccountEAIRate](#accounteairate)

* [AccountHistory](#accounthistory)

* [BlockCurrent](#blockcurrent)

* [BlockHash](#blockhash)

* [BlockHeight](#blockheight)

* [BlockRange](#blockrange)

* [ChaosHistory](#chaoshistory)

* [ChaosNamespaceAll](#chaosnamespaceall)

* [ChaosNamespaceKey](#chaosnamespacekey)

* [NodeStatus](#nodestatus)

* [NodeHealth](#nodehealth)

* [NodeNetInfo](#nodenetinfo)

* [NodeGenesis](#nodegenesis)

* [NodeABCIInfo](#nodeabciinfo)

* [NodeConsensusState](#nodeconsensusstate)

* [NodeList](#nodelist)

* [NodeID](#nodeid)

* [OrderHash](#orderhash)

* [OrderHeight](#orderheight)

* [OrderHistory](#orderhistory)

* [OrderCurrent](#ordercurrent)

* [SystemAll](#systemall)

* [SystemKey](#systemkey)

* [SystemHistoryKey](#systemhistorykey)

* [TransactionByHash](#transactionbyhash)

* [TxPrevalidate](#txprevalidate)

* [TxSubmit](#txsubmit)

* [Version](#version)




---
## AccountByID

### `GET /account/account/:address`

_Returns current state of an account given its address._

Will return an empty result if the account is a valid ID but not on the blockchain.






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "balance": 123000000,
          "validationKeys": [
            "npuba8jadtbbedhhdcad42tysymzpi5ec77vpi4exabh3unu2yem8wn4wv22kvvt24kpm3ghikst"
          ],
          "rewardsTarget": null,
          "incomingRewardsFrom": null,
          "delegationNode": null,
          "lock": null,
          "stake": null,
          "lastEAIUpdate": 0,
          "lastWAAUpdate": 0,
          "weightedAverageAge": 2592000000000,
          "Sequence": 0,
          "settlements": null,
          "settlementSettings": {
            "Period": 0,
            "ChangesAt": null,
            "Next": null
          },
          "validationScript": null
        }
```



---
## AccountsFromList

### `POST /account/accounts`

_Returns current state of several accounts given a list of addresses._

Only returns data for accounts that actively exist on the blockchain.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | []string




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        [
          "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue"
        ]
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue": {
            "balance": 123000000,
            "validationKeys": [
              "npuba8jadtbbedhhdcad42tysymzpi5ec77vpi4exabh3unu2yem8wn4wv22kvvt24kpm3ghikst"
            ],
            "rewardsTarget": null,
            "incomingRewardsFrom": null,
            "delegationNode": null,
            "lock": null,
            "stake": null,
            "lastEAIUpdate": 0,
            "lastWAAUpdate": 0,
            "weightedAverageAge": 2592000000000,
            "Sequence": 0,
            "settlements": null,
            "settlementSettings": {
              "Period": 0,
              "ChangesAt": null,
              "Next": null
            },
            "validationScript": null
          }
        }
```



---
## AccountEAIRate

### `POST /account/eai/rate`

_Returns eai rates for a collection of account information._

Accepts an array of rate requests that includes an address
field; this field may be any string (the account information is not
checked). It returns an array of rate responses, which includes
the address passed so that responses may be correctly correlated
to the input.



_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | []routes.EAIRateRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        [
          {
            "address": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
            "weightedAverageAge": 7776000000000,
            "lock": {
              "noticePeriod": 15552000000000,
              "unlocksOn": null,
              "bonus": 20000000000
            }
          }
        ]
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        [
          {
            "address": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
            "eairate": 6000000
          }
        ]
```



---
## AccountHistory

### `GET /account/history/:accountid`

_Returns the balance history of an account given its address._

The history includes the timestamp, new balance, and transaction ID of each change to the account's balance.
The result is reverse sorted chronologically from the current time, and supports paging by time.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 limit | Query | Maximum number of transactions to return; default=10. | string
 before | Query | Timestamp (ISO 8601) to start looking backwards; default=now. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        [
          {
            "Timestamp": "2018-07-18T20:01:02Z",
            "Balance": 123000000,
            "TxHash": "abc123def456"
          }
        ]
```



---
## BlockCurrent

### `GET /block/current`

_Returns the most recent block in the chain_








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": {
            "block_id": {
              "hash": "",
              "parts": {
                "total": 0,
                "hash": ""
              }
            },
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            }
          },
          "block": {
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            },
            "data": {
              "txs": null
            },
            "evidence": {
              "evidence": null
            },
            "last_commit": null
          }
        }
```



---
## BlockHash

### `GET /block/hash/:blockhash`

_Returns the block in the chain with the given hash._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 blockhash | Path | Hex hash of the block in chain to return. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": {
            "block_id": {
              "hash": "",
              "parts": {
                "total": 0,
                "hash": ""
              }
            },
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            }
          },
          "block": {
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            },
            "data": {
              "txs": null
            },
            "evidence": {
              "evidence": null
            },
            "last_commit": null
          }
        }
```



---
## BlockHeight

### `GET /block/height/:height`

_Returns the block in the chain at the given height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Path | Height of the block in chain to return. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": {
            "block_id": {
              "hash": "",
              "parts": {
                "total": 0,
                "hash": ""
              }
            },
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            }
          },
          "block": {
            "header": {
              "chain_id": "",
              "height": 0,
              "time": "0001-01-01T00:00:00Z",
              "num_txs": 0,
              "total_txs": 0,
              "last_block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "last_commit_hash": "",
              "data_hash": "",
              "validators_hash": "",
              "next_validators_hash": "",
              "consensus_hash": "",
              "app_hash": "",
              "last_results_hash": "",
              "evidence_hash": "",
              "proposer_address": ""
            },
            "data": {
              "txs": null
            },
            "evidence": {
              "evidence": null
            },
            "last_commit": null
          }
        }
```



---
## BlockRange

### `GET /block/range/:first/:last`

_Returns a sequence of block metadata starting at first and ending at last_




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 first | Path | Height at which to begin retrieval of blocks. | int
 last | Path | Height at which to end retrieval of blocks. | int
 noempty | Query | Set to nonblank value to exclude empty blocks | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "last_height": 12345,
          "block_metas": [
            {
              "block_id": {
                "hash": "",
                "parts": {
                  "total": 0,
                  "hash": ""
                }
              },
              "header": {
                "chain_id": "",
                "height": 0,
                "time": "0001-01-01T00:00:00Z",
                "num_txs": 0,
                "total_txs": 0,
                "last_block_id": {
                  "hash": "",
                  "parts": {
                    "total": 0,
                    "hash": ""
                  }
                },
                "last_commit_hash": "",
                "data_hash": "",
                "validators_hash": "",
                "next_validators_hash": "",
                "consensus_hash": "",
                "app_hash": "",
                "last_results_hash": "",
                "evidence_hash": "",
                "proposer_address": ""
              }
            }
          ]
        }
```



---
## ChaosHistory

### `GET /chaos/history/:namespace/:key`

_Returns the history of changes to a value of a single chaos chain variable._

The history includes the block height and the value of each change to the variable.
The result is sorted chronologically.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 namespace | Path | Base-64 (std) text of the namespace, url-encoded. | string
 key | Path | Base-64 (std) name of the variable. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## ChaosNamespaceAll

### `GET /chaos/value/:namespace/all`

_Returns the names and current values of all currently-defined variables in a given namespace on the chaos chain._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 namespace | Path | Base-64 (std) text of the namespace, url-encoded. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## ChaosNamespaceKey

### `GET /chaos/value/:namespace/:key`

_Returns the current value of a single namespaced variable from the chaos chain._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 namespace | Path | Base-64 (std) text of the namespace, url-encoded. | string
 key | Path | Base-64 (std) name of the variable. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## NodeStatus

### `GET /node/status`

_Returns the status of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "node_info": {
            "id": "",
            "listen_addr": "",
            "network": "",
            "version": "",
            "channels": "",
            "moniker": "",
            "other": {
              "amino_version": "",
              "p2p_version": "",
              "consensus_version": "",
              "rpc_version": "",
              "tx_index": "",
              "rpc_address": ""
            }
          },
          "sync_info": {
            "latest_block_hash": "",
            "latest_app_hash": "",
            "latest_block_height": 0,
            "latest_block_time": "0001-01-01T00:00:00Z",
            "catching_up": false
          },
          "validator_info": {
            "address": "",
            "pub_key": null,
            "voting_power": 0
          }
        }
```



---
## NodeHealth

### `GET /node/health`

_Returns the health of the current ndau node and chaos node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "Chaos": {
            "Status": ""
          },
          "Ndau": {
            "Status": ""
          }
        }
```



---
## NodeNetInfo

### `GET /node/net`

_Returns the network information of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "listening": false,
          "listeners": null,
          "n_peers": 0,
          "peers": null
        }
```



---
## NodeGenesis

### `GET /node/genesis`

_Returns the genesis document of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "genesis": null
        }
```



---
## NodeABCIInfo

### `GET /node/abci`

_Returns info on the node's ABCI interface._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "response": {}
        }
```



---
## NodeConsensusState

### `GET /node/consensus`

_Returns the current Tendermint consensus state in JSON_








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "round_state": null,
          "peers": null
        }
```



---
## NodeList

### `GET /node/nodes`

_Returns a list of all nodes._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "nodes": null
        }
```



---
## NodeID

### `GET /node/:id`

_Returns a single node._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 id | Path | the NodeID as a hex string | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "id": "",
          "listen_addr": "",
          "network": "",
          "version": "",
          "channels": "",
          "moniker": "",
          "other": {
            "amino_version": "",
            "p2p_version": "",
            "consensus_version": "",
            "rpc_version": "",
            "tx_index": "",
            "rpc_address": ""
          }
        }
```



---
## OrderHash

### `GET /order/hash/:ndauhash`

_Returns the collection of data from the order chain as of a specific ndau blockhash._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 ndauhash | Path | Hash from the ndau chain. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 0,
          "targetPrice": 0,
          "floorPrice": 0,
          "endowmentSold": 0,
          "totalNdau": 0,
          "priceUnit": ""
        }
```



---
## OrderHeight

### `GET /order/height/:ndauheight`

_Returns the collection of data from the order chain as of a specific ndau block height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 ndauheight | Path | Height from the ndau chain. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 0,
          "targetPrice": 0,
          "floorPrice": 0,
          "endowmentSold": 0,
          "totalNdau": 0,
          "priceUnit": ""
        }
```



---
## OrderHistory

### `GET /order/history`

_Returns an array of data from the order chain at periodic intervals over time, sorted chronologically._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 limit | Query | Maximum number of values to return; default=100, max=1000. | string
 period | Query | Duration between samples (ex: 1d, 5m); default=1d. | string
 before | Query | Timestamp (ISO 8601) to end (exclusive); default=now. | string
 after | Query | Timestamp (ISO 8601) to start (inclusive); default=before-(limit*period). | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        []
```



---
## OrderCurrent

### `GET /order/current`

_Returns current order chain data for key parameters._

Returns current order chain information for 5 parameters:
* Market price
* Target price
* Floor price
* Total ndau sold from the endowment
* Total ndau in circulation







_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 16.85,
          "targetPrice": 17,
          "floorPrice": 2.57,
          "endowmentSold": 291900000000000,
          "totalNdau": 314159300000000,
          "priceUnit": "USD"
        }
```



---
## SystemAll

### `GET /system/all`

_Returns the names and current values of all currently-defined system variables._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## SystemKey

### `GET /system/:key`

_Returns the current value of a single system variable._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 key | Path | Name of the system variable. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## SystemHistoryKey

### `GET /system/history/:key`

_Returns the history of changes to a value of a system variable._

The history includes the timestamp, new value, and transaction ID of each change to the value.
The result is reverse sorted chronologically from the current time, and supports paging by time.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 key | Path | Name of the system variable. | string
 limit | Query | Maximum number of values to return; default=10. | string
 before | Query | Timestamp (ISO 8601) to start looking backwards; default=now. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TransactionByHash

### `GET /transaction/:txhash`

_Returns a transaction from the blockchain given its tx hash._

Transaction hash must be URL query-escaped






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "Tx": null
        }
```



---
## TxPrevalidate

### `POST /tx/prevalidate`

_Prevalidates a transaction._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxJSON




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {
          "data": "base64 tx data"
        }
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "fee_napu": 10,
          "err": "only set if an error occurred"
        }
```



---
## TxSubmit

### `POST /tx/submit`

_Submits a transaction._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxJSON




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {
          "data": "base64 tx data"
        }
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "hash": "123abc34099f"
        }
```



---
## Version

### `GET /version`

_Delivers version information_








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "ChaosVersion": "",
          "ChaosSha": "",
          "NdauVersion": "v1.2.3",
          "NdauSha": "3123abc35",
          "Network": "ndau mainnet"
        }
```
