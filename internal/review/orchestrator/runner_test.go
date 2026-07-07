package orchestrator

import (
	"context"
	"testing"

	"aicr/internal/model"
	"aicr/internal/provider"
)

type fakeProvider struct {
	patch    provider.PatchSet
	summary  string
	inline   []provider.InlineComment
	statuses []provider.ReviewStatus
}

func (f *fakeProvider) GetPatchSet(context.Context, string, int) (provider.PatchSet, error) {
	return f.patch, nil
}

func (f *fakeProvider) PublishSummary(_ context.Context, _ string, _ int, comment provider.SummaryComment) error {
	f.summary = comment.Body
	return nil
}

func (f *fakeProvider) PublishInline(_ context.Context, _ string, _ int, _ string, comment provider.InlineComment) error {
	f.inline = append(f.inline, comment)
	return nil
}

func (f *fakeProvider) UpdateStatus(_ context.Context, _ string, _ string, status provider.ReviewStatus) error {
	f.statuses = append(f.statuses, status)
	return nil
}

type fakeReviewModel struct{}

func (fakeReviewModel) Review(context.Context, model.ReviewModelInput, model.ModelPolicy) (model.ReviewModelOutput, error) {
	return model.ReviewModelOutput{
		Summary:   "issue found",
		RiskLevel: "medium",
		Findings: []model.Finding{{
			Severity:      "P1",
			Category:      "correctness",
			File:          "main.go",
			Line:          2,
			Message:       "bug",
			EvidenceLevel: "A",
			Confidence:    0.9,
		}},
	}, nil
}

func TestRunnerCompletesReviewFlow(t *testing.T) {
	provider := &fakeProvider{patch: provider.PatchSet{
		Provider: "github",
		Repo:     "acme/widgets",
		Number:   5,
		HeadSHA:  "abc123",
		Files: []provider.ChangedFile{{
			Path:     "main.go",
			Status:   "modified",
			Patch:    "@@ -1 +1 @@\n-old\n+new",
			Contents: "package main\n",
		}},
	}}
	runner := NewRunner(RunnerOptions{
		Provider: provider,
		Model:    fakeReviewModel{},
	})

	run, err := runner.Run(context.Background(), Request{Repo: "acme/widgets", Number: 5})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if run.State != StateCompleted {
		t.Fatalf("run state = %s, want completed", run.State)
	}
	if len(provider.inline) != 1 {
		t.Fatalf("inline comments = %d, want 1", len(provider.inline))
	}
	if provider.summary == "" {
		t.Fatal("summary was not published")
	}
	if len(provider.statuses) != 2 {
		t.Fatalf("statuses = %d, want 2", len(provider.statuses))
	}
}
