package orchestrator

import (
	"context"
	"fmt"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/agent"
	"aicr/internal/review/aggregate"
	"aicr/internal/review/chunker"
	reviewcontext "aicr/internal/review/context"
	"aicr/internal/review/publish"
)

// Request 是执行一次 review run 所需的最小输入。
type Request struct {
	Repo   string // 仓库全名，例如 "acme/widgets"。
	Number int    // PR/MR 编号。
}

// RunnerOptions 是 Runner 依赖的外部能力。
type RunnerOptions struct {
	Provider provider.ReviewerProvider // GitHub/GitLab 等 provider adapter。
	Model    model.ReviewModelClient   // 模型网关 client。
}

// Runner 串起完整的 review pipeline。
//
// 它不直接知道 GitHub API 细节，也不直接知道模型 HTTP 协议细节；
// 这些都通过接口注入进来，方便测试和后续替换。
type Runner struct {
	provider provider.ReviewerProvider
	model    model.ReviewModelClient
}

// NewRunner 创建 review runner。
func NewRunner(opts RunnerOptions) *Runner {
	return &Runner{
		provider: opts.Provider,
		model:    opts.Model,
	}
}

// Run 执行一次完整审查。
//
// 当前 MVP 是同步执行：拉 diff -> 拆 unit -> 构建 context -> 调模型 ->
// 聚合 findings -> 发布 summary/inline/status。后续可以把每个 unit 的模型调用
// 并行化，或拆到独立 worker 中。
func (r *Runner) Run(ctx context.Context, req Request) (*Run, error) {
	run := NewRun(fmt.Sprintf("%s:%d", req.Repo, req.Number))
	if err := run.Transition(StateQueued); err != nil {
		return run, err
	}
	if err := run.Transition(StateLoadingMetadata); err != nil {
		return run, err
	}
	if err := run.Transition(StateBuildingPatchset); err != nil {
		return run, err
	}
	// 从 provider 拉取 PR/MR diff，转换成统一的 PatchSet。
	patch, err := r.provider.GetPatchSet(ctx, req.Repo, req.Number)
	if err != nil {
		_ = run.Transition(StateFailedRetryable)
		run.FailureReason = err.Error()
		return run, err
	}
	if err := run.Transition(StateChunking); err != nil {
		return run, err
	}
	// MVP 先按文件拆 review unit，后续可扩展成按 hunk/symbol/DAG 拆分。
	units := chunker.ChunkPatchSet(patch)

	if err := run.Transition(StateReviewing); err != nil {
		return run, err
	}
	runtime := agent.NewRuntime(r.model)
	var outputs []model.ReviewModelOutput
	var findings []model.Finding
	for _, unit := range units {
		// Context Pack 是给模型的证据包，避免模型只看孤立 diff。
		pack := reviewcontext.BuildContextPack(unit, reviewcontext.BuildOptions{})
		output, err := runtime.Review(ctx, pack)
		if err != nil {
			_ = run.Transition(StateFailedRetryable)
			run.FailureReason = err.Error()
			return run, err
		}
		outputs = append(outputs, output)
		findings = append(findings, output.Findings...)
	}

	if err := run.Transition(StateAggregating); err != nil {
		return run, err
	}
	// 聚合阶段做去重、证据等级过滤和评论预算控制，降低 review 噪音。
	aggregation := aggregate.Aggregate(findings, aggregate.Options{InlineBudget: 5})

	if err := run.Transition(StatePublishing); err != nil {
		return run, err
	}
	// 先标记 pending，给用户一个“审查正在进行/已进入发布阶段”的状态反馈。
	_ = r.provider.UpdateStatus(ctx, patch.Repo, patch.HeadSHA, provider.ReviewStatus{State: "pending", Description: "AI review running"})
	summary, risk := summarize(outputs)
	if err := r.provider.PublishSummary(ctx, patch.Repo, patch.Number, provider.SummaryComment{Body: publish.BuildSummary(publish.SummaryInput{
		OverallSummary: summary,
		RiskLevel:      risk,
		Aggregation:    aggregation,
	})}); err != nil {
		_ = run.Transition(StateFailedRetryable)
		run.FailureReason = err.Error()
		return run, err
	}
	for _, comment := range publish.BuildInlineComments(aggregation) {
		if err := r.provider.PublishInline(ctx, patch.Repo, patch.Number, patch.HeadSHA, comment); err != nil {
			_ = run.Transition(StateFailedRetryable)
			run.FailureReason = err.Error()
			return run, err
		}
	}
	if err := r.provider.UpdateStatus(ctx, patch.Repo, patch.HeadSHA, provider.ReviewStatus{State: "success", Description: "AI review completed"}); err != nil {
		_ = run.Transition(StateFailedRetryable)
		run.FailureReason = err.Error()
		return run, err
	}
	if err := run.Transition(StateCompleted); err != nil {
		return run, err
	}
	return run, nil
}

// summarize 把多个 review unit 的模型输出合并成一个总评摘要。
//
// MVP 使用简单拼接；后续可以引入专门的 aggregator model 做二次总结。
func summarize(outputs []model.ReviewModelOutput) (string, string) {
	if len(outputs) == 0 {
		return "No review units found.", "low"
	}
	var summary string
	risk := "low"
	for _, output := range outputs {
		if output.Summary != "" {
			if summary != "" {
				summary += "\n"
			}
			summary += output.Summary
		}
		if output.RiskLevel != "" && output.RiskLevel != "low" {
			risk = output.RiskLevel
		}
	}
	return summary, risk
}
