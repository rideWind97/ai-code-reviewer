package main

import (
	"log/slog"
	"os"

	"aicr/internal/config"
)

// main 是后台 worker 进程的入口。
//
// 当前文件还是占位骨架：它先验证配置能正常读取，后续会在这里接入持久化队列，
// 消费 webhook 入队的 review job，并调用 orchestrator.Runner 执行完整审查流程。
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	// 这里暂时只打印关键配置是否存在，避免在日志里输出真实密钥。
	slog.Info("aicr-worker ready", "database_configured", cfg.DatabaseURL != "", "model", cfg.Model.Name)
}
