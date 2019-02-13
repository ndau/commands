# `verify-pre-eai`: Pre-EAI verification of genesis account data

The mandate here is fairly simple:

- read the genesis file
- for each account, verify the following fields by comparing genesis to blockchain:
    - balance
    - last EAI date
