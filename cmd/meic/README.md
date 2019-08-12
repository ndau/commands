# Multi-Exchange Issuance Coordinator

We need a way to coordinate activity on the ndau blockchain with an arbitrary number of exchanges. Coordination happens in two directions:

- When target price sales occur on the exchange, `Issue` transactions must be created and dispatched to the blockchain.
- Whenever the sales offers change, all exchanges must then update their target price sales offers appropriately.

## General Architecture

This software is divided into two major components. The **Issuance Update System** (IUS) is the heart: it maintains a persistent websocket connection from the signing service, which it uses to sign `Issue` transactions. The **Order Tracking System** (OTS) is an interface implemented by custom software for each exchange. Its role is to hide the exchange's implementation details from the IUS.

The IUS doesn't know or care whether an OTS implementation has websocket connections to its exchange, offers a webhook API, polls the exchange API, or something else. Implementations are broadly free to do whatever they want to perform updates to and from their exchanges as close to realtime as possible. However, they must only ever communicate with the ndau blockchain via the IUS.

Note: though each OTS implementation runs in its own goroutine and has a clear separation of concerns from the IUS, this software is still built as a single executable. It might be desirable in the future to split the IUS and OTS implementations into independent executables, possibly running on different hardware, communicating via (web)sockets. For now we have chosen not to do this: we aren't adding new exchanges fast enough that coordinating their additions is a problem, and it's just simpler to write this as a monolith with channels than as a constellation of microservices with websockets.

## Bitmart OTS Notes

### Api Keys

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
