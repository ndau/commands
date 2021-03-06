# `ndsh` -- the Ndau Shell

The `ndau` tool is very useful and powerful, but its design is oriented toward
a development environment in which convenience is more important than security.
This means that all its internal state, including private keys, is persisted
in plain text on the user's hard drive. For operational use, we need something
more security-oriented.

`ndsh` fills that need: it stores nothing outside of volatile memory, making it
safe to run in secure operational environments.

## Features

### Implemented

- is a shell
    - has a prompt
    - can exit to surrounding shell with `exit` or `quit`
- launch with a `--net=X` argument, where `X` can be `main`, `test`, `dev`, `local`, or any URL. Default to `main`.
- specify commands to execute on launch, and post-execution exit policy
- enter a 12-word phrase after launch: it isn't exposed to your shell history
- automatically asynchronously discover accounts for a given phrase
- manually add undiscovered accounts by derivation path
- refer to accounts by nicknames or minimal suffixes
- list known accounts and nicknames
- manually add nicknamed "foreign" accounts by address
    - or specify them from the command line (use `-c`)
- view account details
- do most things the ndau tool can do:
    - accounts
        - create new account, return address and derivation path
        - set validation rules (with recovery from recovery service as applicable)
        - create a child account
        - update status from blockchain
        - add, remove validation script
        - perform CRUD operations on arbitrary validation keys
        - change recourse period
        - delegate
        - lock
        - notify
        - stake
        - set rewards destination
        - register node
        - claim node reward
        - send `CreditEAI` tx
        - closeout account into another account, transfering out all ndau
    - rfe
    - issue
    - transfer
    - transfer and lock
    - get and set system variables
    - get version information
    - get summary/sib information
    - nominate node rewards
    - record price
    - command validation change
- staged mode for transactions:
    - override any field (via JSON representation)
    - add signatures from arbitrary private keys
    - add signatures directly from certain hardware keys
    - just emit the signable bytes of the current state
    - serialize the JSON out, or `send` to send to the blockchain
- certain commands (`summary`, `tx`, `view`) have `--jq` option to filter the output

## Conventions

`ndsh` expects that every `Command` implement a safe, idempotent `-h` flag which
displays some help about that command.
