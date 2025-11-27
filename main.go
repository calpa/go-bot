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
	loadEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	botClient := newBot()

	// call methods.SetWebhook if needed

	go botClient.StartWebhook(ctx)

	mux := newMux(botClient)

	srv := newServer(mux)

	startServer(ctx, srv)

	// call methods.DeleteWebhook if needed
}

func loadEnv() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		println("Error loading .env file:", err.Error())
	}
}

func newBot() *bot.Bot {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithWebhookSecretToken(os.Getenv("TELEGRAM_WEBHOOK_SECRET_TOKEN")),
	}

	botClient, _ := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	return botClient
}

func newMux(b *bot.Bot) *http.ServeMux {
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
			fmt.Fprintf(w, "failed to set webhook to %s: %v", webhookURL, err)
			return
		}

		if !ok {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "Telegram did not accept webhook URL: %s", webhookURL)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "webhook set to %s", webhookURL)
	})

	return mux
}

func newServer(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
}

func startServer(ctx context.Context, srv *http.Server) {
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
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	fmt.Println("Received update: ", update.Message.Text)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Echo: " + update.Message.Text,
	})
}
