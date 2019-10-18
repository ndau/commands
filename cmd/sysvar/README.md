# `sysvar`

System variables require specialized encoding to work properly; the precise
encoding depends on the type of the variable. Things are _much_ easier when
the variable type is known.

This utility just encodes sysvars of known type into base64 strings suitable
for inserting into the JSON representation of a `SetSysvar` tx.
