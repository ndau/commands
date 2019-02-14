# `verify`: verification of genesis account data

## `pre-eai.py`

The mandate here is fairly simple:

- read the genesis file
- for each account, verify the following fields by comparing genesis to blockchain:
    - balance
    - last EAI date
