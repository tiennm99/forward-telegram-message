import asyncio

from telethon import TelegramClient
from telethon import events

from config import *

client = TelegramClient(phone, api_id, api_hash)


@client.on(events.NewMessage(chats=[source_group_id]))
async def forward_message(event):
    try:
        await client.forward_messages(target_group_id, event.message)
    except Exception as e:
        print(f"Exception on forward_message: {str(e)}")


async def main():
    while True:
        try:
            print("Starting...")
            await client.start(phone)
            print("Bot is running.")
            await client.run_until_disconnected()
        except Exception as e:
            print(f"Exception on main: {str(e)}. Restarting...")


if __name__ == '__main__':
    asyncio.run(main())
