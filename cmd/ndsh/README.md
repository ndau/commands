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

### Planned

- launch with a `--net=X` argument, where `X` can be `main`, `test`, `dev`, `local`, or any URL. Default to `main`.
- enter a 12-word phrase after launch: it isn't exposed to your shell history
- automatically asynchronously discover accounts for a given phrase
- `await-discovery`: wait for discovery to finish, show progress
- manually add undiscovered accounts by derivation path
- manually add nicknamed "foreign" accounts by address
    - or specify them from the command line
- refer to accounts by nicknames or minimal suffixes
- list known accounts and nicknames
- do most things the ndau tool can do:
    - accounts
        - create new account, return address and derivation path
        - claim account
        - claim child account
        - closeout account into another account, transfering out all ndau
        - add, remove, list, or edit validation script (specify `.chasm` or `.chbin` file)
        - perform CRUD operations on arbitrary validation keys
        - update status from blockchain
        - change recourse period
        - delegate
        - send `CreditEAI` tx
        - lock
        - notify
        - set rewards target
        - stake
        - register node
        - claim node reward
    - transfer
    - transfer and lock
    - rfe
    - issue
    - nominate node rewards
    - command validation change
    - record price
    - get and set system variables
    - get sib information
    - get version information
- `prepare` mode for transactions:
    - override any field
    - add signatures from arbitrary private keys
    - add signatures directly from certain harware keys
    - just emit the signable bytes of the current state
    - when complete, `write` to serialize the JSON out, or `send` to send to the blockchain
- arrow-pageable history for non-secret commands
- customize prompt by (externally) setting `NDSH_PS1` with useful substitutions (?)
- tab-completion (?)
- job control (?)
