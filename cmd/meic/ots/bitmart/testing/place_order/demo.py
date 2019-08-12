#!/usr/bin/env python3
import requests
import hmac
import hashlib
from urllib.parse import urlencode

url = "https://openapi.bitmart.com/v2/orders"

secret = "api_secret"  # normally you'd use a real one

# notice: the price and amount should be formatted as *.******, not scientific notation
data = {"symbol": "BMX_ETH", "amount": 1.5, "price": 1.234, "side": "buy"}

# notice: the parameters should be sorted in alphabetical order to produce the correct
# signature
sorted_data = sorted(data.items(), key=lambda d: d[0], reverse=False)
print("sorted data:", sorted_data)
message = str(urlencode(sorted_data))
print("message:    ", message)
signed_message = hmac.new(
    secret.encode("utf-8"), message.encode("utf-8"), hashlib.sha256
).hexdigest()
print("signed msg: ", signed_message)

req = requests.Request(
    "POST", url, json=data, headers={"x-bm-signature": signed_message}
)
prep = req.prepare()
print("req headers:", prep.headers)
print("req body:   ", prep.body)

