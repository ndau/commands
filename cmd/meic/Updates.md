# Updates to directories

Here I'll describe updates to the directories in the bitmart repo:

## bitmart

- main.go - Updates to support Kent's new signature service API

## bitmart/api


I've updated a few files in the api directory to provide some loggging info, but no major functionality changes.



## bitmart/testing/auth_rest

- bitmart_prod_example.py - example of getting orders history and trade history from Bitmart production API - this needs to be invoked with the "prod.apikey.json" file

- bitmart_uat_orders.py - example of placing orders on the Bitmart UAT API - this needs to be invoked with the "test.apikey.json" file

- ws_trades.py - example Bitmart trades websocket subscription

- kryptono_example.py - example Kryptono account details for UAT account - this needs to be invoked with the "kryptono.apikey.json" file

## bitmart/testing/mock_bitmart

- sigconfig.toml - I've updated this file to conform to Kent's new format for signature service config files.  They keys in this file should work for signing an Issue TX for your localnet, but will not work for testnet or mainet due to Yubikey reqs

- test.sh - I've updated this file with keys and commands that should successfully run the issuance service against the mock Bitmart API and generate an Issue TX for localnet, will not work for testnet/mainnet due to comment above

## bitmart/testing/uat_bitmart

- sigconfig.toml/test.sh - Updated these files to support running the issuance service against the Bitmart UAT account/API

## bitmart/testing/prod_bitmart

- sigconfig.toml/test.sh - Updated these files to support running the issuance service against the Bitmart production account/API

## bitmart/testing/orders

- main.go - WIP to test comparing blockchain issuance number with Bitmart issuance number from pending orders on issuance account

## bitmart/testing/trades

- main.go - WIP to test comparing blockchain issuance number with Bitmart issuance number from trade history on issuance account

