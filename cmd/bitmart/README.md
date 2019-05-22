# Bitmart Integrations

We need an automatic service that can watch the Bitmart exchange and automatically
issue appropriate `Issue` transactions on the `ndau` chain when appropriate.

## General Architecture

- **Signing Service**: [external oneiro software](https://github.com/oneiro-ndev/recovery/tree/master/cmd/signer). Connects to a websocket server, and then listens for signature requests. On receipt, signs and returns them. Note: structurally this is a client, even though behaviorally this is a server. This is a bit unusual, but it simplifies the security requirements.

    The signing service will connect to the Bitmart Integration.

- **Bitmart Integration**: this software. Polls the Bitmart REST API for new trades from a particular account. For each batch of new trades from this account, it calculates the total ndau traded from those trades which were sales. It then creates an `Issue` tx, has it signed by the signing service, and submits it to the ndau blockchain.

## Api Keys

Bitmart requires that certain API requests be authenticated. This is documented [here](https://github.com/bitmartexchange/bitmart-official-api-docs/blob/master/rest/authenticated/oauth.md).

To keep track of that for our application, create a file whose extension is `apikey.json` within the `bitmart` folder. It will be gitignored, protecting these secrets. Using the example data from the bitmart auth document, `example.apikey.json` must look like:

```json
{
    "access": "6591f7c2491db0a23a1d8ad6911c825e",
    "secret": "8c08d9d5c3d15b105dbddaf96e427ac6",
    "memo": "mymemo"
}
```

- `access` is the API key
- `secret` is the API secret
- `memo` is the human-friendly name the user provided when the API key was created

It is occasionally necessary to override the endpoint used for a particular key. This is mainly useful for testing. If necessary, just add an `endpoint` field to the json file, like:

```json
{
    "access": "6591f7c2491db0a23a1d8ad6911c825e",
    "secret": "8c08d9d5c3d15b105dbddaf96e427ac6",
    "memo": "mymemo",
    "endpoint": "https://bm-htf-v2-testing-d8pvw98nl.bitmart.com"
}
```
