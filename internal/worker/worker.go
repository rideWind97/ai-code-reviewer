package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aicr/internal/queue"
	"aicr/internal/review/orchestrator"
	"aicr/internal/reviewjob"
)

// ReviewRequest 是 worker 从队列 payload 中解析出来的审查请求。
type ReviewRequest = reviewjob.Request

// Runner 是 worker 依赖的审查执行接口。
//
// 这里单独定义小接口，避免 worker 直接依赖 orchestrator.Runner 的具体类型。
type Runner interface {
	Run(context.Context, ReviewRequest) error
}

// Worker 从队列中取出 review job 并交给 Runner 执行。
type Worker struct {
	queue  queue.Queue
	runner Runner
}

// LoopOptions 控制 worker 常驻循环。
type LoopOptions struct {
	IdleDelay time.Duration // 队列为空时等待多久再轮询。
}

// New 创建 worker。
func New(q queue.Queue, runner Runner) *Worker {
	return &Worker{queue: q, runner: runner}
}

// ProcessOne 处理一个队列任务。
//
// 返回 ok=false 表示队列为空；调用方可以 sleep 后继续轮询。
func (w *Worker) ProcessOne(ctx context.Context) (bool, error) {
	job, ok, err := w.queue.Dequeue(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if job.Type != "review" {
		return false, fmt.Errorf("unsupported job type: %s", job.Type)
	}
	var req ReviewRequest
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		return false, fmt.Errorf("decode review job: %w", err)
	}
	return true, w.runner.Run(ctx, req)
}

// RunLoop 持续消费队列，直到 context 被取消或遇到错误。
func (w *Worker) RunLoop(ctx context.Context, opts LoopOptions) error {
	if opts.IdleDelay <= 0 {
		opts.IdleDelay = time.Second
	}
	for {
		ok, err := w.ProcessOne(ctx)
		if err != nil {
			return err
		}
		if ok {
			continue
		}
		timer := time.NewTimer(opts.IdleDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// OrchestratorRunner 是 orchestrator.Runner 的最小接口。
type OrchestratorRunner interface {
	Run(context.Context, orchestrator.Request) (*orchestrator.Run, error)
}

// ProviderRouterRunner 根据 job 中的 provider 选择对应 orchestrator runner。
type ProviderRouterRunner struct {
	runners map[string]OrchestratorRunner
}

// NewProviderRouterRunner 创建 provider 路由 runner。
func NewProviderRouterRunner(runners map[string]OrchestratorRunner) *ProviderRouterRunner {
	return &ProviderRouterRunner{runners: runners}
}

// Run 把 worker.ReviewRequest 转换成 orchestrator.Request。
func (r *ProviderRouterRunner) Run(ctx context.Context, req ReviewRequest) error {
	runner := r.runners[req.Provider]
	if runner == nil {
		return fmt.Errorf("missing runner for provider: %s", req.Provider)
	}
	_, err := runner.Run(ctx, orchestrator.Request{Repo: req.Repo, Number: req.Number})
	return err
}
