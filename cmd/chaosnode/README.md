# `chaosnode`: ABCI application implementing a chaos chain node

## Usage

- Terminal 1:

    ```sh
    $ chaosnode
    ```

- Terminal 2:

    ```sh
    $ tendermint node
    ```

- Terminal 3:

    Set and read values from the chaos blockchain using the `chaos` tool:

    ```sh
    $ chaos set -k key -v value
    ```

    and so forth. See `chaos` tool documentation for more details.