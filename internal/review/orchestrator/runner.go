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

type Request struct {
	Repo   string
	Number int
}

type RunnerOptions struct {
	Provider provider.ReviewerProvider
	Model    model.ReviewModelClient
}

type Runner struct {
	provider provider.ReviewerProvider
	model    model.ReviewModelClient
}

func NewRunner(opts RunnerOptions) *Runner {
	return &Runner{
		provider: opts.Provider,
		model:    opts.Model,
	}
}

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
	patch, err := r.provider.GetPatchSet(ctx, req.Repo, req.Number)
	if err != nil {
		_ = run.Transition(StateFailedRetryable)
		run.FailureReason = err.Error()
		return run, err
	}
	if err := run.Transition(StateChunking); err != nil {
		return run, err
	}
	units := chunker.ChunkPatchSet(patch)

	if err := run.Transition(StateReviewing); err != nil {
		return run, err
	}
	runtime := agent.NewRuntime(r.model)
	var outputs []model.ReviewModelOutput
	var findings []model.Finding
	for _, unit := range units {
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
	aggregation := aggregate.Aggregate(findings, aggregate.Options{InlineBudget: 5})

	if err := run.Transition(StatePublishing); err != nil {
		return run, err
	}
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
