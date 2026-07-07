package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/lib/pq"

	"aicr/internal/config"
	"aicr/internal/model"
	"aicr/internal/provider/github"
	"aicr/internal/provider/gitlab"
	"aicr/internal/queue"
	"aicr/internal/review/orchestrator"
	"aicr/internal/worker"
)

// main 是后台 worker 进程的入口。
//
// 它连接 PostgreSQL 队列，消费 webhook 入队的 review job，并调用 orchestrator.Runner。
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		slog.Error("AICR_DATABASE_URL is required for aicr-worker")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	reviewQueue := queue.NewPostgres(&queue.SQLDB{DB: db})
	modelClient := model.NewOpenAICompatibleClient(model.ClientConfig{
		BaseURL: cfg.Model.BaseURL,
		APIKey:  cfg.Model.APIKey,
		Model:   cfg.Model.Name,
	})
	runner := worker.NewProviderRouterRunner(map[string]worker.OrchestratorRunner{
		"github": orchestrator.NewRunner(orchestrator.RunnerOptions{
			Provider: github.NewClient(github.ClientOptions{Token: cfg.GitHub.Token}),
			Model:    modelClient,
		}),
		"gitlab": orchestrator.NewRunner(orchestrator.RunnerOptions{
			Provider: gitlab.NewClient(gitlab.ClientOptions{BaseURL: cfg.GitLab.BaseURL, Token: cfg.GitLab.Token}),
			Model:    modelClient,
		}),
	})
	reviewWorker := worker.New(reviewQueue, runner)

	slog.Info("aicr-worker started", "model", cfg.Model.Name)
	if err := reviewWorker.RunLoop(context.Background(), worker.LoopOptions{}); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
