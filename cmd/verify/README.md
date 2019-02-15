# `verify`: verification of genesis account data

## `pre-eai.py`

The mandate here is fairly simple:

- read the genesis file
- for each account, verify the following fields by comparing genesis to blockchain:
    - balance
    - last EAI date

## `post-eai.py`

The main goal here is to consume the spreadsheet CSV and emit a new one,
containing all its current data plus some new columns:

- credited EAI
- currency seats in date order (ordinal)
