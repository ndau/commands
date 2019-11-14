# `nh`: Noms History

`nh` is a tool to inspect the noms history, customized for the ndau use-case.
It allows for searches over higher-level concepts, such as accounts and addresses,
rather than noms' low-level `Value`s.

## Usage

By default, `nh` simply summarizes the database:

```sh
$ go run ./cmd/nh data/noms/
state summary:
    2497 accounts
       5 nodes
```
