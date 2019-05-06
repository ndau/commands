import requests
import json
import hmac

AUTH_API = "https://openapi.bitmart.com/v2/authentication"


def create_sha256_signature(key, message):
    return hmac.digest(key, message, "sha256").hex()


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
    print(data)
    response = requests.post(AUTH_API, data=data)  # note: _not_ JSON!
    return response.json()


def get_keys(path):
    with open(path) as f:
        return json.load(f)


if __name__ == "__main__":
    import sys

    path = sys.argv[1]
    keys = get_keys(path)
    access_token = get_access_token(keys)
    print(access_token)
