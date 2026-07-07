package main

import (
	"log/slog"
	"net/http"
	"os"

	"aicr/internal/config"
	"aicr/internal/httpserver"
	"aicr/internal/provider/github"
	"aicr/internal/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	reviewQueue := queue.NewMemory()
	handler := httpserver.New(httpserver.Options{
		GitHubWebhook: github.WebhookHandler(github.WebhookOptions{
			Secret: cfg.GitHub.WebhookSecret,
			Queue:  reviewQueue,
		}),
	})

	slog.Info("starting aicr-server", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
