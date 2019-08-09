# `base36`: Quickly convert to/from base36 values

For example: all numbers are stored at the noms level as base36-encoded strings,
because noms' native integer support is atrocious and this compacts things fairly
well.

If you want to get the noms' head's current height, `cd` into the noms dir, then:

```sh
$ noms show .::ndau.value.Height | base36
26267
```

The tool reads its value from its first argument if one is provided, or stdin otherwise.
With the `-E` flag, it encodes base10 numbers to base36.
