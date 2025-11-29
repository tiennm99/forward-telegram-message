# forward-telegram-message

A Telegram message forwarder that forwards messages from a source chat to a target chat.

This project is implemented in Go using the gotd/td Telegram client library.

## Features

- Forward messages from one Telegram chat to another
- Automatic reconnection on disconnect
- Proper error handling and logging
- Environment variable configuration

## Requirements

- Go 1.24 or higher
- Telegram API credentials (visit https://my.telegram.org/ to get them)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/tiennm99/forward-telegram-message.git
cd forward-telegram-message
```

2. Install dependencies:
```bash
go mod tidy
```

## Configuration

Create a `.env` file in the project root with the following content:

```env
TG_PHONE=+1234567890
API_ID=123456789
API_HASH=0123456789abcdef0123456789abcdef
SOURCE_GROUP_ID=-123456789
TARGET_GROUP_ID=-123456789
```

Replace the values with your Telegram credentials and chat IDs.

## Usage

1. Install dependencies:
```bash
go mod tidy
```

2. Run the application:
```bash
go run .
```

The application will connect to Telegram and start forwarding messages from the source chat to the target chat.

## The Original Python Version

The original Python version of this project can be found at the `feature/python` branch.
