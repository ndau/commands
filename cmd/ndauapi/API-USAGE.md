ndau API Usage - Best Practices
===

## Finding a node endpoint
Applications should never rely on the availability of a single blockchain node. The Oneiro development team
runs a set of validator nodes with public addresses, but they are no different than any other node for
API purposes.
### _Retrieving the list of Oneiro nodes_
The URLs of Oneiro's nodes are stored in a JSON service description on Amazon S3.
```
https://s3.us-east-2.amazonaws.com/ndau-json/services.json
```
That service should always be queried first to obtain a list of API addresses if an application wants
to submit transactions and queries to an Oneiro node.
### _Checking a node's availability_
The API endpoint `/health` must always be queried before attempting to use any node. Nodes may be
offline at any time for backups or maintenance. The `/health` endpoint will simply return "OK" if the
node is operational and available.
### _Catching up_
When a node comes back online it will be behind the other nodes and will need to catch up to the current
block on the ndau blockchain. The `/node/status` endpoint returns JSON that includes the node's sync info:
```
"sync_info": {
    "latest_block_hash": "05B25C3FCF60B87FB25B8FDC12AAAB43D3C0031F6D4F214461365FDC8DF361A9",
    "latest_app_hash": "6920C05074A5779FB480870FB96AD40EAD874DB2",
    "latest_block_height": 530346,
    "latest_block_time": "2022-04-06T17:10:04.310050494Z",
    "catching_up": false
  }
```
If `sync_info.catching_up` is true, the node is not yet reporting the current state of the blockchain.
Transactions may be submitted to it, but any query will report out-of-date information, including
the current block height of the blockchain.
## Special addresses
ndau addresses used for special purposes are stored as system variables. The `/system/all` endpoint
returns a list and values of all system variables, and `/system/get/<variable name>` will return the
value of a single variable.
### Staking addresses
ndau supports staking for a variety of purposes. Each purpose is governed by a set of staking rules
associated with an ndau address. Staking addresses may change, and should never be hardcoded. They
should be retrieved from the appropriate system variable whenever used.
### *Validator node registration rules*
The validator node registration rules address can be retrieved from the `NodeRulesAccountAddress`:
```
/system/get/NodeRulesAccountAddress
```
