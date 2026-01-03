# Token Hunter Bot

A Telegram bot token generator and tester written in Go.

## Deployment on Koyeb/Render

1. **GitHub**: Upload this code to a GitHub repository.
2. **Koyeb**:
   - Select "GitHub" as the deployment source.
   - Build Command: `go build -o token-hunter .`
   - Run Command: `./token-hunter`
   - Set Environment Variables:
     - `BOT_TOKEN`: Your Telegram bot token.
     - `CHAT_ID`: Your Telegram chat ID.
     - `PORT`: 8080 (or your preferred port).
3. **Workers**: You can change the number of workers using the `-workers` flag (e.g., `./token-hunter -workers 500`).

## Strategy

The bot uses a hybrid strategy:
- 50% Workers: Target common bot ID ranges (5,000,000,000 to 9,000,000,000).
- 50% Workers: Pure random generation.

## Local Development

```bash
export BOT_TOKEN="your_token"
export CHAT_ID="your_chat_id"
go run main.go
```
