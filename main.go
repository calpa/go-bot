package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: b.WebhookHandler(),
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
