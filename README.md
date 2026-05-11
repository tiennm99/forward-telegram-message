# forward-telegram-message

Mirror Telegram messages between two chats using Telethon — runs as a user-account, not a bot.

## Quick start

```bash
cp .env.example .env
# Fill in PHONE, API_ID, API_HASH, SOURCE_GROUP_ID, TARGET_GROUP_ID
pip install -r requirements.txt
python main.py
```

Get `API_ID` and `API_HASH` from [my.telegram.org](https://my.telegram.org).

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `PHONE` | Yes | Your Telegram phone number (e.g. `+84912345678`) |
| `API_ID` | Yes | App API ID from my.telegram.org |
| `API_HASH` | Yes | App API hash from my.telegram.org |
| `SOURCE_GROUP_ID` | Yes | ID of the source chat/group to mirror from |
| `TARGET_GROUP_ID` | Yes | ID of the destination chat/group to forward into |

## Notes

- Runs as a **user account** (Telethon), not a bot — requires first-time interactive login to generate a session file.
- Forwards both text messages and media (photos, documents, videos).
- Re-uses the session file on subsequent runs; no re-auth needed.
- `SOURCE_GROUP_ID` and `TARGET_GROUP_ID` can be negative (supergroup IDs) or positive (private chat IDs).

## License

Apache-2.0 — see [LICENSE](LICENSE).
