package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		println("Error loading .env file:", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithWebhookSecretToken(os.Getenv("TELEGRAM_WEBHOOK_SECRET_TOKEN")),
	}

	b, _ := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)

	// call methods.SetWebhook if needed

	go b.StartWebhook(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	})

	mux.Handle("/telegram/webhook", b.WebhookHandler())

	mux.HandleFunc("/set-webhook", func(w http.ResponseWriter, r *http.Request) {
		baseURL := os.Getenv("WEBHOOK_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8000"
		}

		baseURL = strings.TrimRight(baseURL, "/")
		webhookURL := baseURL + "/telegram/webhook"

		ok, err := b.SetWebhook(r.Context(), &bot.SetWebhookParams{
			URL: webhookURL,
		})

		w.Header().Set("Content-Type", "text/plain")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to set webhook to %s: %v", webhookURL, err)))
			return
		}

		if !ok {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(fmt.Sprintf("Telegram did not accept webhook URL: %s", webhookURL)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("webhook set to %s", webhookURL)))
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	go func() {
		println("HTTP server listening on :8000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			println("HTTP server error:", err.Error())
		}
	}()

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		println("HTTP server shutdown error:", err.Error())
	}

	// call methods.DeleteWebhook if needed
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}
