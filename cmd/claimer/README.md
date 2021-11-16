# `claimer`

Listen for node reward nominations, and claim them if they come to us.

When the ndau node processes a `NominateNodeReward` transaction, it checks
its configuration: if the config includes a URL for a node reward webhook,
the node sends a POST request whose body is a simple JSON package:

```json
{
		"random": 97823457892,     // the random number which produced this result
		"winner": "ndnsomeaddress" // the ndau address of the winning node
}
```

The intent is that node operators will configure their nodes to point to a
web service that they operate, and that this service knows sufficient keys to
generate a `ClaimNodeReward` transaction. `claimer` is a sample implementation of such a webhook service. Other approaches to watching for and claiming node rewards are possible.

## Configuration and Usage

For each node under management, the claimer must be able to sign a `ClaimNodeReward` transaction with the node's private key(s). It must also be able to submit this signed transaction to a node. This is all configured in a TOML configuration file, called by convention `claimer_conf.toml`. An example of such a file:

```toml
node_api = "http://localhost:3030"

[nodes]
ndah7rywpmw3geashr6yrayzrw8i527t56k3xhvr6yy638v4 = [
    "npvtayjadtcbib5s3exhq5vb77ijkck4reema2apcvqnjaj3hbfxrssdqnmdjuq5i8z3sij64qby9843piarmef768tbb5uewdqw9k447gaqrr8vwtqke33prc62",
]
```

Run the claimer. If `claimer_conf.toml` is not in the same directory as the executable, specify its location on the command line:

```sh
$ ./claimer --config-path=cmd/claimer/claimer_conf.toml
Tue Jul 16 11:05:49 CEST 2019
INFO[0000] using API address                             bin=claimer node address="http://localhost:3030"
INFO[0000] qty keys known per known node                 bin=claimer ndah7rywpmw3geashr6yrayzrw8i527t56k3xhvr6yy638v4=1
INFO[0000] server listening                              port=8080 rootpath=/
```

By default the claimer listens on port 8080. Use the command line argument `port=portnum` to select a different port. The `=` sign in the assignment is _not_ optional.

You can also specify these parameters via a environment variables of the same names:

```sh
$ CONFIG_PATH=cmd/claimer/claimer_conf.toml PORT=9999 ./claimer
Tue Jul 16 11:09:56 CEST 2019
INFO[0000] using API address                             bin=claimer node address="http://localhost:3030"
INFO[0000] qty keys known per known node                 bin=claimer ndah7rywpmw3geashr6yrayzrw8i527t56k3xhvr6yy638v4=1
INFO[0001] server listening                              port=9999 rootpath=/
```
The claimer will report the actions it takes and their results when its webhook is called by the node.
```I
NFO[0010] REQ                                           bin=claimer code=200 host="localhost:8080" len=19 method=POST remoteAddr="[::1]:57925" rootpath=/ took="108.661µs" ua=curl/7.64.1 uri=/claim_winner {"dispatched":true}/
INFO[0010] successfully claimed node reward              bin=claimer nodeURL="http://localhost:3030" rootpath=/ winnerAddress=ndah7rywpmw3geashr6yrayzrw8i527t56k3xhvr6yy638v4
```
