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
NDAUAPI_NDAU_RPC_URL=http://127.0.0.1:31001 ./ndauapi
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
        "NDAUAPI_NDAU_RPC_URL": "http://127.0.0.1:31001"
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

* [DEPRECATEDAccountEAIRate](#deprecatedaccounteairate)

* [AccountHistory](#accounthistory)

* [AccountList](#accountlist)

* [AccountCurrencySeats](#accountcurrencyseats)

* [BlockBefore](#blockbefore)

* [BlockCurrent](#blockcurrent)

* [BlockHash](#blockhash)

* [BlockHeight](#blockheight)

* [BlockRange](#blockrange)

* [BlockTransactions](#blocktransactions)

* [BlockDateRange](#blockdaterange)

* [NodeStatus](#nodestatus)

* [NodeHealth](#nodehealth)

* [NodeNetInfo](#nodenetinfo)

* [NodeGenesis](#nodegenesis)

* [NodeABCIInfo](#nodeabciinfo)

* [NodeConsensusState](#nodeconsensusstate)

* [NodeList](#nodelist)

* [NodeID](#nodeid)

* [DEPRECATEDOrderCurrent](#deprecatedordercurrent)

* [OrderHeight](#orderheight)

* [OrderHistory](#orderhistory)

* [PriceInfo](#priceinfo)

* [StateDelegates](#statedelegates)

* [SystemAll](#systemall)

* [SysvarGet](#sysvarget)

* [SysvarSet](#sysvarset)

* [SysvarHistory](#sysvarhistory)

* [AccountEAIRate](#accounteairate)

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
```
        {
          "balance": 123000000,
          "validationKeys": [
            "npuba8jadtbbedhhdcad42tysymzpi5ec77vpi4exabh3unu2yem8wn4wv22kvvt24kpm3ghikst"
          ],
          "validationScript": null,
          "rewardsTarget": null,
          "incomingRewardsFrom": null,
          "delegationNode": null,
          "lock": null,
          "lastEAIUpdate": "2000-01-01T00:00:00Z",
          "lastWAAUpdate": "2000-01-01T00:00:00Z",
          "weightedAverageAge": "1m",
          "sequence": 0,
          "stake_rules": null,
          "costakers": null,
          "holds": null,
          "recourseSettings": {
            "period": "t0s",
            "changes_at": null,
            "next": null
          },
          "currencySeatDate": null,
          "parent": null,
          "progenitor": null
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
            "validationScript": null,
            "rewardsTarget": null,
            "incomingRewardsFrom": null,
            "delegationNode": null,
            "lock": null,
            "lastEAIUpdate": "2000-01-01T00:00:00Z",
            "lastWAAUpdate": "2000-01-01T00:00:00Z",
            "weightedAverageAge": "1m",
            "sequence": 0,
            "stake_rules": null,
            "costakers": null,
            "holds": null,
            "recourseSettings": {
              "period": "t0s",
              "changes_at": null,
              "next": null
            },
            "currencySeatDate": null,
            "parent": null,
            "progenitor": null
          }
        }
```



---
## DEPRECATEDAccountEAIRate

### `POST /account/eai/rate`

_This call is deprecated -- please use /system/eai/rate._






_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        null
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        null
```



---
## AccountHistory

### `GET /account/history/:address`

_Returns the balance history of an account given its address._

The history includes the timestamp, new balance, and transaction ID of each change to the account's balance.
The result is sorted chronologically.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 address | Path | The address of the account for which to return history | string
 after | Query | The block height after which results should start. | string
 limit | Query | The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100 | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "Items": [
            {
              "Balance": 123000000,
              "Timestamp": "2018-07-10T20:01:02Z",
              "TxHash": "L4hD20bp7w4Hi19vpn46wQ",
              "Height": 0
            }
          ]
        }
```



---
## AccountList

### `GET /account/list`

_Returns a list of account IDs._

This returns a list of every account on the blockchain, sorted
alphabetically. A maximum of 10000 accounts can be returned in a single
request. The results are sorted by address.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 after | Query | The address after which (sorted alphabetically) results should start. | string
 limit | Query | The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100 | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "NumAccounts": 1,
          "FirstIndex": 1,
          "After": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "NextAfter": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "Accounts": [
            "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue"
          ]
        }
```



---
## AccountCurrencySeats

### `GET /account/currencyseats`

_Returns a list of ndau 'currency seats', which are accounts containing more than 1000 ndau._

The ndau currency seats are accounts containing more than 1000 ndau. The seniority of
a currency seat is determined by how long it has been above the 1000 threshold, so this endpoint
also sorts the result by age (oldest first). It does not return detailed account information.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 limit | Query | The max number of items to return (default=3000) | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "NumAccounts": 1,
          "FirstIndex": 1,
          "After": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "NextAfter": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "Accounts": [
            "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue"
          ]
        }
```



---
## BlockBefore

### `GET /block/before/:height`

_Returns a (possibly filtered) sequence of block metadata for blocks of height less than last._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Path | Blocks of this height and greater will not be returned. | int
 filter | Query | Set to 'noempty' to exclude empty blocks. | string
 after | Query | The block height after which no more results should be returned. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
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
                "version": {
                  "block": 0,
                  "app": 0
                },
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
## BlockCurrent

### `GET /block/current`

_Returns the most recent block in the chain_








_**Produces:**_ `[application/json]`


_**Writes:**_
```
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
              "version": {
                "block": 0,
                "app": 0
              },
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
              "version": {
                "block": 0,
                "app": 0
              },
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
```
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
              "version": {
                "block": 0,
                "app": 0
              },
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
              "version": {
                "block": 0,
                "app": 0
              },
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
```
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
              "version": {
                "block": 0,
                "app": 0
              },
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
              "version": {
                "block": 0,
                "app": 0
              },
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
```
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
                "version": {
                  "block": 0,
                  "app": 0
                },
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
## BlockTransactions

### `GET /block/transactions/:height`

_Returns transaction hashes for a given block. These can be used to fetch data for individual transactions._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Path | Height of the block in chain containing transactions. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        [
          "L4hD20bp7w4Hi19vpn46wQ"
        ]
```



---
## BlockDateRange

### `GET /block/daterange/:first/:last`

_Returns a sequence of block metadata starting at first date and ending at last date_




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 first | Path | Timestamp (ISO 3339) at which to begin (inclusive) retrieval of blocks. | string
 last | Path | Timestamp (ISO 3339) at which to end (exclusive) retrieval of blocks. | string
 noempty | Query | Set to nonblank value to exclude empty blocks | string
 after | Query | The timestamp after which results should start (use the last value from the previous page). | string
 limit | Query | The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100 | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
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
                "version": {
                  "block": 0,
                  "app": 0
                },
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
## NodeStatus

### `GET /node/status`

_Returns the status of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "node_info": {
            "protocol_version": {
              "p2p": 0,
              "block": 0,
              "app": 0
            },
            "id": "",
            "listen_addr": "",
            "network": "",
            "version": "",
            "channels": "",
            "moniker": "",
            "other": {
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

_Returns the health of the current node by doing a simple test for connectivity and response._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
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
```
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
```
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
```
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
```
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
```
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




---
## DEPRECATEDOrderCurrent

### `GET /order/current`

_This is an obsolete format. Please use /price/current instead._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        null
```



---
## OrderHeight

### `GET /price/height/:height`

_Returns the collection of price data as of a specific ndau block height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Path | Height from the ndau chain. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "marketPrice": 0,
          "targetPrice": 0,
          "totalIssued": 0,
          "totalNdau": 0,
          "totalSIB": 0,
          "sib": 0
        }
```



---
## OrderHistory

### `GET /price/history`

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
```
        []
```



---
## PriceInfo

### `GET /price/current`

_Returns current price data for key parameters._

Returns current price information:
* Market price
* Target price
* Total ndau issued from the endowment
* Total ndau in circulation
* Total SIB burned
* Current SIB in effect







_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "marketPrice": 1234000000000,
          "targetPrice": 5678000000000,
          "totalIssued": 291900000000000,
          "totalNdau": 314159300000000,
          "totalSIB": 12300000000,
          "sib": 9876543210
        }
```



---
## StateDelegates

### `GET /state/delegates`

_Returns the current collection of delegate information._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        ""
```



---
## SystemAll

### `GET /system/all`

_Returns the names and current values of all currently-defined system variables._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        ""
```



---
## SysvarGet

### `GET /system/get/:sysvars`

_Return the names and current values of some currently definted system variables._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 sysvars | Path | A comma-separated list of system variables of interest. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        ""
```



---
## SysvarSet

### `POST /system/set/:sysvar`

_Returns a transaction which sets a system variable._

The body of the request accepts JSON and heuristically transforms
it into the data format used internally on the blockchain. Do not use any sort
of wrapping object. The correct structure of the object to send depends on
the system variable in question.

Returns the JSON encoding of a SetSysvar transaction. It is the caller's
responsibility to update this transaction with appropriate sequence and
signatures and then send it at the normal endpoint (/tx/submit/setsysvar).


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 sysvar | Path | The name of the system variable to return | string




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        null
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## SysvarHistory

### `GET /system/history/:sysvar`

_Returns the value history of a system variable given its name._

The history includes the height and value of each change to the system variable.
The result is sorted chronologically.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 sysvar | Path | The name of the system variable for which to return history | string
 after | Query | The block height after which results should start. | string
 limit | Query | The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100 | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "history": [
            {
              "height": 12345,
              "value": "VmFsdWU="
            }
          ]
        }
```



---
## AccountEAIRate

### `POST /system/eai/rate`

_Returns eai rates for a collection of account information._

Accepts an array of rate requests that includes an address
field; this field may be any string (the account information is not
checked). It returns an array of rate responses, which includes
the address passed so that responses may be correctly correlated
to the input.

It accepts a timestamp, which will be used to adjust WAA in the
event the account is locked and has a non-nil "unlocksOn" value.
If the timestamp field is omitted, the current time is used.

EAIRate in the response is an integer equal to the fractional EAI
rate times 10^12.



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
            "weightedAverageAge": "3m",
            "lock": {
              "noticePeriod": "6m",
              "unlocksOn": null,
              "bonus": 20000000000
            },
            "at": "2018-07-10T20:01:02Z"
          }
        ]
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        [
          {
            "address": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
            "eairate": 60000000000
          }
        ]
```



---
## TransactionByHash

### `GET /transaction/:txhash`

_Returns a transaction from the blockchain given its tx hash._








_**Produces:**_ `[application/json]`


_**Writes:**_
```
        {
          "Tx": null
        }
```



---
## TxPrevalidate

### `POST /tx/prevalidate/:txtype`

_Prevalidates a transaction (tells if it would be accepted and what the transaction fee will be._

Transactions consist of JSON for any defined transaction type (see submit).


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | *ndau.Lock




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {
          "target": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "period": "1m",
          "sequence": 1234,
          "signatures": null
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

### `POST /tx/submit/:txtype`

_Submits a transaction._

Transactions consist of JSON for any defined transaction type. Valid transaction names are: change-recourse-period, changerecourseperiod, changeschema, changesettlementperiod, changevalidation, claim, claim-child, claimaccount, claimchildaccount, claimnodereward, commandvalidatorchange, create-child, create-child-account, createchildaccount, crediteai, crp, cvc, delegate, issue, lock, nnr, nominatenodereward, notify, record-price, recordprice, registernode, releasefromendowment, rfe, set-validation, setrewardsdestination, setstakerules, setsysvar, setv, setvalidation, ssv, stake, transfer, transferandlock, unregisternode, unstake


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | *ndau.Lock




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {
          "target": "ndamgmmntjwhq37gi6rwpazy4fka6zgzix55x85kkhepvuue",
          "period": "1m",
          "sequence": 1234,
          "signatures": null
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
```
        {
          "ChaosVersion": "",
          "ChaosSha": "",
          "NdauVersion": "v1.2.3",
          "NdauSha": "3123abc35",
          "Network": "ndau mainnet"
        }
```
