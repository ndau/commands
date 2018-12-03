# KeyAddr

The file main.go contains code designed to be built with GopherJS to create a JS implementation of the same functions for key generation and manipulation.

To build, use `go generate`.

## JavaScript

All functions generate promises, so .


### To generate a master key:

`masterkey = address.NewPrivateMaster('random data');`

The master key is a private key.

### To generate a public key from a master key:

`publickey = key.Public()`

### To generate a child at index N from any kind of key (public or private):

`ch = key.Child(n)`

Note that the child is of the same type (public or private) as the parent.

Public child N is paired with Private child N, no matter how they are derived. So:

`private.Child(27).Public()` is the same key as `private.Public().Child(27)`, and both are the public keys associated with the private key called `private.Child(27)`.

### To generate a hardened child:

`hch = key.HardenedChild(n)`

A hardened child is a child that cannot be used to make more children.


