package main

import (
	"log/slog"
	"net/http"
	"os"

	"aicr/internal/config"
	"aicr/internal/httpserver"
	"aicr/internal/provider/github"
	"aicr/internal/provider/gitlab"
	"aicr/internal/queue"
)

// main 是 HTTP 服务进程的入口。
//
// 当前 MVP 只负责两件事：
// 1. 暴露 /healthz，方便部署平台或本地开发确认服务存活。
// 2. 暴露 /webhooks/github，接收 GitHub PR 事件并写入 review 队列。
//
// 真正的 review 执行逻辑放在 internal/review/orchestrator，避免入口文件变复杂。
func main() {
	// Go 服务通常启动时先读取配置；这里的配置来自环境变量。
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	// MVP 先使用内存队列，方便本地跑通链路。
	// 后续接 PostgreSQL/SQS/Redis 时，只需要替换 queue.Queue 的实现。
	reviewQueue := queue.NewMemory()
	handler := httpserver.New(httpserver.Options{
		GitHubWebhook: github.WebhookHandler(github.WebhookOptions{
			Secret: cfg.GitHub.WebhookSecret,
			Queue:  reviewQueue,
		}),
		GitLabWebhook: gitlab.WebhookHandler(gitlab.WebhookOptions{
			SecretToken: cfg.GitLab.SecretToken,
			Queue:       reviewQueue,
		}),
	})

	slog.Info("starting aicr-server", "addr", cfg.Addr)
	// ListenAndServe 会阻塞当前 goroutine，直到服务异常退出。
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
