# Bitmart Integrations

We need an automatic service that can watch the Bitmart exchange and automatically
issue appropriate `Issue` transactions on the `ndau` chain when appropriate.

## General Architecture

- **Signing Service**: external oneiro software. Connects to a websocket server, and then listens for signature requests. On receipt, signs and returns them. Note: structurally this is a client, even though behaviorally this is a server. This is a bit unusual, but it simplifies the security requirements.

    An instance of the signing service will be a client of the

- **Bitmart Integration**: this software. It subscribes to messages about NDAU sales from the primary sales account on Bitmart. For each of these, it generates an appropriate `Issue` tx and sends it to the signing service. On getting it back, it sends it to the ndau chain.

    It is a client of the ndau chain, and of the

- **Bitmart Websocket Service**: external software run by bitmart. Provides near-real-time push updates about transactions of interest.

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

