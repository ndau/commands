#!/usr/bin/env python3
import requests
import json
import hmac
import hashlib
from urllib.parse import urlencode
import decimal

# AUTH_API = "https://openapi.bitmart.com/v2/authentication"
API = "https://bm-htf-v2-testing-d8pvw98nl.bitmart.com/v2/"
SYMBOL = "BMX_USDT"
AUTH_API = API + "authentication"
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


def get_access_token(keys):
    api_key = keys["access"]
    api_secret = keys["secret"]
    memo = keys["memo"]

    message = api_key + ":" + api_secret + ":" + memo
    data = {
        "grant_type": "client_credentials",
        "client_id": api_key,
        "client_secret": create_sha256_signature(
            api_secret.encode("utf-8"), message.encode("utf-8")
        ),
    }
    print("request data = ")
    print(data)
    response = requests.post(AUTH_API, data=data)  # note: _not_ JSON!
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
    access_token = get_access_token(keys)
    print("access token = ")
    print(access_token)
    time = get_time()

    # timestamp is in milliseconds and the authorization header is "Bearer " + token
    headers = {"X-BM-TIMESTAMP": time["server_time"], "X-BM-AUTHORIZATION": "Bearer " + access_token["access_token"]}

    response = requests.get(WALLET_API, headers=headers)

    print(response.text)

    response = requests.get(ORDERS_API + ORDERS_PARAMS_PEND, headers=headers)

    print("pending orders = " + response.text)

    # response = requests.get(ORDERS_API + ORDERS_PARAMS_SUCC, headers=headers)

    # print("succ orders = " + response.text)


    response = requests.get(TRADES_API + TRADES_HIST, headers=headers)

    print("succ trades = " + response.text)


    buy_data = {"symbol": "BMX_ETH","amount": 2,"price" : .099,"side" : "buy"}
    signed_message = create_signed_message(buy_data, keys["secret"])

    data = json.dumps(buy_data)

    headers = {"X-BM-TIMESTAMP": time["server_time"], "X-BM-AUTHORIZATION": "Bearer " + access_token["access_token"], "X-BM-SIGNATURE": signed_message, "Content-Type": "application/json"}

    response = requests.post(ORDERS_API, data=data, headers=headers)

    print(response.text)

    sell_data = {"symbol": "BMX_USDT","amount": 1000,"price" : .0334,"side" : "sell"}
    signed_message = create_signed_message(sell_data, keys["secret"])

    data = json.dumps(sell_data)

    headers = {"X-BM-TIMESTAMP": time["server_time"], "X-BM-AUTHORIZATION": "Bearer " + access_token["access_token"], "X-BM-SIGNATURE": signed_message, "Content-Type": "application/json"}

    response = requests.post(ORDERS_API, data=data, headers=headers)

    print(response.text)

    # timestamp is in milliseconds and the authorization header is "Bearer " + token
    headers = {"X-BM-TIMESTAMP": time["server_time"], "X-BM-AUTHORIZATION": "Bearer " + access_token["access_token"]}

    response = requests.get(ORDERS_API + ORDERS_PARAMS_PEND, headers=headers)

    print("pending orders = " + response.text)

    # response = requests.get(WALLET_API, headers=headers)

#    print(response.text)

    # response = requests.get(ORDERS_API + ORDERS_PARAMS_SUCC, headers=headers)

    # print("succ orders = " + response.text)


