#!/usr/bin/env python3

import asyncio
import json
import websockets

BITMART = "wss://openws.bitmart.com"


async def ping():
    print("creating websocket connection to bitmart...")
    async with websockets.connect(BITMART) as ws:
        print("sending 'ping' subscription...")
        await ws.send('{"subscribe":"ping"}')
        print("waiting for response...")
        data = await ws.recv()

    try:
        print(json.loads(data))
    except Exception as e:
        print("json error:", e)
        print(data)


def main():
    asyncio.run(ping())


if __name__ == "__main__":
    main()
