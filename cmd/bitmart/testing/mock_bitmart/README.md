# `mock_bitmart`: mock out the bitmart API for testing

It turns out that we need to implement only a few endpoints in order to support the issuance service:

- authentication at `/v2/authentication`
- trade history at `/v2/trades`
- order detail at `/v2/orders/:entrust_id`

This program therefore just fakes up a bunch of internally-consistent trades and orders, and serves them locally at the appropriate paths. External services can then connect to those endpoints for testing.

## Notes
- This is a static service: the orders and trades lists will not change over the runtime of the program
- No attempt is made to authenticate incoming requests
