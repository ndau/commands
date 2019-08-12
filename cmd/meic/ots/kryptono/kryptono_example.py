#!/usr/bin/env python3
import requests
import json
import hmac
import hashlib
from urllib.parse import urlencode
import decimal

API = "https://testenv1.kryptono.exchange/k/api/v2/"
SYMBOL = "XND_USDT"
ACCOUNT_DETAILS_API = API + "account/details"
ACCOUNT_BALANCES_API = API + "account/balances"
TIME_API = API + "time"
WALLET_API = API + "wallet"
ORDERS_API = API + "orders"
TRADES_API = API + "trades"
ORDERS_PARAMS_PEND = "?symbol=" + SYMBOL + "&status=5&offset=0&limit=100"
ORDERS_PARAMS_SUCC = "?symbol=" + SYMBOL + "&status=3&offset=0&limit=100"
TRADES_HIST = "?symbol=" + SYMBOL + "&limit=10&offset=0"

def create_sha256_signature(key, message):
    return hmac.digest(key, message, "sha256").hex()

def get_time():
    response = requests.get(TIME_API)
    return response.json()


def get_keys(path):
    with open(path) as f:
        return json.load(f)

def create_signed_message(data, secret):
    # notice: the parameters should be sorted in alphabetical order to produce the correct
    # signature
    print(data)
    sorted_data = sorted(data.items(), key=lambda d: d[0], reverse=False)
    print("sorted data:", sorted_data)
    message = str(urlencode(sorted_data))
    print("message:    ", message)
    secret = keys["secret"]
    signed_message = hmac.new(
        secret.encode("utf-8"), message.encode("utf-8"), hashlib.sha256
    ).hexdigest()
    print("signed msg: ", signed_message)
    return signed_message

if __name__ == "__main__":
    import sys

    path = sys.argv[1]
    keys = get_keys(path)
    time = get_time()
#    print(f'time = {time}')

    data = f'timestamp={time["server_time"]}'
    # timestamp is in milliseconds and the authorization header is "Bearer " + token
    headers = {"Authorization": keys["access"], 
            "Signature": create_sha256_signature(keys["secret"].encode("utf-8"), data.encode("utf-8"))
            }

    response = requests.get(ACCOUNT_DETAILS_API + "?" + data, headers=headers)

    print(response.text)

    response = requests.get(ACCOUNT_BALANCES_API + "?" + data, headers=headers)

    print(response.text)


