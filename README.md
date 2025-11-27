# Go Telegram Echo Bot

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=git&repository=https://github.com/calpa/go-bot&branch=main&name=go-bot)

A simple Telegram bot written in Go. It runs as a webhook-based bot, echoes back any text message it receives (with an `Echo: ` prefix), exposes a health/index route, and provides a helper endpoint to set the Telegram webhook URL.

## Features

- **Echo bot**: Replies with `Echo: <your message>` to incoming text messages.
- **HTTP server on port 8000**:
  - `/` → returns `Hello World` (plain text).
  - `/telegram/webhook` → Telegram webhook endpoint used by the bot.
  - `/set-webhook` → helper route to call Telegram's `setWebhook` for this app.
- **Graceful shutdown**: Listens for `Ctrl+C` / `SIGINT` and shuts down the HTTP server cleanly.

## Requirements

- Go 1.20+ (or compatible with your environment)
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- (Optional) A public HTTPS URL for webhooks (Koyeb, ngrok, etc.)

## Configuration

The application uses environment variables, typically loaded from a `.env` file via `godotenv`.

Required variables:

- `TELEGRAM_BOT_TOKEN` – your bot token from BotFather
- `TELEGRAM_WEBHOOK_SECRET_TOKEN` – secret token used to validate Telegram webhook calls
- `WEBHOOK_BASE_URL` – **optional** base URL used by `/set-webhook` to build the full webhook URL
  - If not set, it defaults to `http://localhost:8000`.

Example `.env` file:

```env
TELEGRAM_BOT_TOKEN=123456789:example-token
TELEGRAM_WEBHOOK_SECRET_TOKEN=some-secret-string
WEBHOOK_BASE_URL=https://your-app-url.example.com
```

## Running locally

1. Install dependencies:

```bash
go mod tidy
```

2. Ensure your `.env` file is present in the project root.

3. Run the bot:

```bash
go run main.go
```

You should see output similar to:

```text
HTTP server listening on :8000
```

Then:

- Open `http://localhost:8000/` → you should see `Hello World`.
- `Ctrl+C` in the terminal will terminate the server gracefully.

## Setting the Telegram webhook

The bot exposes a helper route to configure the webhook from your running app.

1. Ensure `WEBHOOK_BASE_URL` is set appropriately:
   - For local testing with a tunnel (e.g. ngrok), set it to your public URL, e.g. `https://<ngrok-id>.ngrok.io`.
   - For Koyeb, use your Koyeb app URL such as `https://<your-app-name>-<region>.koyeb.app`.

2. With the app running, open in your browser or via `curl`:

```text
GET /set-webhook
```

Examples:

```bash
curl http://localhost:8000/set-webhook
# or after deployment
curl https://<your-app-url>/set-webhook
```

The response will show whether Telegram accepted the webhook and the URL it was set to.

## Project structure

- `main.go` – contains:
  - `loadEnv` – loads environment variables from `.env`.
  - `newBot` – creates the Telegram `*bot.Bot` with the default handler and webhook secret.
  - `newMux` – builds the HTTP mux and routes (`/`, `/telegram/webhook`, `/set-webhook`).
  - `newServer` – constructs the `http.Server` on `:8000`.
  - `startServer` – starts the HTTP server and performs graceful shutdown on context cancel.
  - `handler` – Telegram update handler that logs and echoes incoming text messages.

## Deploying to Koyeb

You can deploy this repository to [Koyeb](https://www.koyeb.com/) and run it as a web service.

1. Push this repository to GitHub (or another supported Git provider).
2. Click the **Deploy to Koyeb** button at the top of this README, which opens the Koyeb dashboard.
3. In Koyeb, choose your repository and branch.
4. Set the following environment variables in the Koyeb service settings:
   - `TELEGRAM_BOT_TOKEN`
   - `TELEGRAM_WEBHOOK_SECRET_TOKEN`
   - `WEBHOOK_BASE_URL` – set to your Koyeb app URL, e.g. `https://<your-app-name>-<region>.koyeb.app`
5. Deploy the service.

After the service is running on Koyeb:

1. Visit `https://<your-app-url>/set-webhook` once to register the webhook with Telegram.
2. Send a message to your bot in Telegram and it should respond with `Echo: <your message>`.

## Notes

- The `.gitignore` ignores common Go build artifacts and `.env` files so your secrets are not committed.
- Adjust ports, paths, or logging as needed for your environment.
