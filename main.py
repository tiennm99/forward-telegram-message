import getpass

from telethon import events
from telethon.errors import SessionPasswordNeededError
from telethon.sync import TelegramClient

from config import *

# client = TelegramClient(phone, api_id, api_hash)

# client.connect()
# if not client.is_user_authorized():
#     client.send_code_request(phone)
#     try:
#         client.sign_in(code=input('Enter code: '))
#     except SessionPasswordNeededError:
#         client.sign_in(password=getpass.getpass())

# client.start(phone)

with TelegramClient('session_name', api_id, api_hash) as client:
    @client.on(events.NewMessage(chats=source_group_id))
    async def handler(event):
        await client.forward_messages(destination_group_id, event.message)

    client.run_until_disconnected()
