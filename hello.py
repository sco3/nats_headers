#!/usr/bin/env python3

import asyncio
from nats.aio.client import Client as NATS


async def run():
    nc = NATS()

    # Connect to the NATS server
    await nc.connect("nats://127.0.0.1:4222")
    subj = "test.subject"

    # Callback function for message handling
    async def message_handler(msg):
        subject = msg.subject
        data = msg.data.decode()
        headers = msg.headers  # Corrected: Use msg.headers instead of msg.header
        print(f"Received a message on subject '{subject}': {data}")
        print(f"Headers: {headers}")

    # Subscribe to a subject
    await nc.subscribe(subj, cb=message_handler)

    print("Server listening on subject 'test.subject'...")

    # Keep sending messages to the subject with headers
    count = 0
    while True:
        message = f"Hello, NATS with Headers {count}!"

        # Define headers
        headers = {
            "asdf": "asdf",
            "jkl": "jkl",
        }

        # Publish the message to the subject with headers
        await nc.publish(subj, message.encode(), headers=headers)

        count += 1
        await asyncio.sleep(1)


if __name__ == "__main__":
    asyncio.run(run())
