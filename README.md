# forward-telegram-message

Forward every new message from one Telegram chat to another with Go and
[gotd/td](https://github.com/gotd/td). It runs as a **user account**, not a
bot, so it can read any chat the account is in. Text and media (photos,
documents, videos) are forwarded with Telegram's "Forwarded from" header.

## How to use

1. Clone this repository and install [Go](https://go.dev/dl/):

```bash
git clone https://github.com/tiennm99/forward-telegram-message.git
```

2. Create a new Telegram app at [https://my.telegram.org](https://my.telegram.org) and get the `api_id` and `api_hash`.
3. Copy `.env.example` to `.env` and fill in the values.
4. Run the forwarder:

```bash
go run ./cmd/forward
```

The first run asks for the Telegram login code (and the two-step
verification password, if the account has one), then stores the session in
`session.json`. Later runs reuse it and **do not** prompt again. If the login
fails (no answer on stdin, a wrong code or password, or Telegram rejecting the
phone), the forwarder exits instead of retrying, so it does not keep
requesting new codes. Run the first login in an interactive terminal.

The forwarder runs until interrupted (Ctrl+C). When the connection drops or
Telegram returns an error, it reconnects after 5 seconds. A message that fails
to forward is reported and skipped.

Console output keeps the Python version's wording (`Starting...`,
`Bot is running.`, `Exception on forward_message: ...`,
`Exception on main: ... Restarting...`) on stdout. Errors that stop the
forwarder (bad config, failed login, unknown chat) go to stderr with exit
status 1.

## Configuration

Set these in `.env` or in the environment:

| Variable | Description |
|----------|-------------|
| `PHONE` | Your Telegram phone number in international format (e.g. `+84912345678`) |
| `API_ID` | App API ID from my.telegram.org |
| `API_HASH` | App API hash from my.telegram.org |
| `SOURCE_GROUP_ID` | Chat to forward from |
| `TARGET_GROUP_ID` | Chat to forward into |

Chat IDs use the usual marked form: `-100<id>` for supergroups and channels,
`-<id>` for basic groups, and the other person's user ID (a positive number)
for private chats. The account must already be in both chats: the forwarder
looks them up in its chat list at startup and stops if either is missing.

Messages sent while the forwarder is running are forwarded in order,
including ones the account itself sends to the source chat. Messages sent
during a reconnect are caught up once the connection is back; messages sent
while the process is not running are not. Service messages such as joins and
pins are skipped.

## Session file

`session.json` grants full access to the Telegram account. It is written with
`0600` permissions and ignored by Git; delete it to force a clean re-login.
Sessions from the earlier Python version (`*.session`) are not reused, so the
first Go run asks for a login code once.

## Build a binary

The code is pure Go, so a static binary builds without cgo:

```bash
CGO_ENABLED=0 go build -o bin/ ./cmd/forward
```

## Development

```bash
go vet ./...
go test ./...
```

Tests run offline; Telegram calls are not exercised.

## License

Apache-2.0 — see [LICENSE](LICENSE).
