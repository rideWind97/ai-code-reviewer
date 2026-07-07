package main

import (
	"log/slog"
	"os"

	"aicr/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	slog.Info("aicr-worker ready", "database_configured", cfg.DatabaseURL != "", "model", cfg.Model.Name)
}
