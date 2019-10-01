#!/usr/bin/env python3
import json
from websocket import create_connection
import ssl


if __name__ == "__main__":

    ws = create_connection("wss://engine-test.kryptono.exchange/ws/v1/ht/KNOW_USDT", sslopt={"cert_reqs": ssl.CERT_NONE})

#    ws.send('{"subscribe":"trade","symbol":"XND_USDT","precision":4}')

    result =  json.loads(ws.recv())

#    print(f'number of trades: {len(result["data"]["trades"])}')

    print(result)

    result =  json.loads(ws.recv())

    # print(f'number of trades: {len(result["data"]["trades"])}')

    print(result)

    ws.close()



