# based on https://github.com/bitmartexchange/bitmart-official-api-docs/
#                             blob/master/websocket/ping.md

import json
from websocket import create_connection
from contextlib import closing


def main():
    print("creating websocket connection to bitmart...")
    with closing(create_connection("wss://openws.bitmart.com")) as ws:
        print("sending 'ping' subscription...")
        ws.send('{"subscribe":"ping"}')
        print("waiting for response...")
        data = ws.recv()

    try:
        print(json.loads(data))
    except Exception as e:
        print("json error:", e)
        print(data)


if __name__ == "__main__":
    main()
