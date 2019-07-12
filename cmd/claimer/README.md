# `claimer`

Listen for node reward nominations, and claim them if they come to us.

When the ndau node processes a `NominateNodeReward` transaction, it checks
its configuration: if the config includes a URL for a node reward webhook,
it sends a POST request whose body is a simple JSON package:

```json
{
		"random": 97823457892,    // the random number which produced this result
		"winner": "ndnsomeaddress", // the ndau address of the winning node
}
```

The intent is that node operators will configure their nodes to point to a
web service that they operate, and that this service knows sufficient keys to
generate a `ClaimNodeReward` transaction.

`claimer` is the canonical Ndev implementation of such a service.
