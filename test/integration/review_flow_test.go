package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/provider/github"
	"aicr/internal/queue"
	"aicr/internal/review/orchestrator"
)

func TestWebhookToReviewRunnerFlow(t *testing.T) {
	q := queue.NewMemory()
	webhook := github.WebhookHandler(github.WebhookOptions{Secret: "secret", Queue: q})
	body := []byte(`{
		"action": "opened",
		"repository": {"full_name": "acme/widgets"},
		"pull_request": {
			"number": 5,
			"draft": false,
			"head": {"sha": "abc123"},
			"base": {"sha": "def456"}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature("secret", body))
	rec := httptest.NewRecorder()

	webhook.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("webhook status = %d", rec.Code)
	}
	if len(q.Jobs()) != 1 {
		t.Fatalf("jobs = %d, want 1", len(q.Jobs()))
	}

	provider := &flowProvider{patch: provider.PatchSet{
		Provider: "github",
		Repo:     "acme/widgets",
		Number:   5,
		HeadSHA:  "abc123",
		BaseSHA:  "def456",
		Files: []provider.ChangedFile{{
			Path:     "main.go",
			Status:   "modified",
			Patch:    "@@ -1 +1 @@\n-old\n+new",
			Contents: "package main\n",
		}},
	}}
	runner := orchestrator.NewRunner(orchestrator.RunnerOptions{
		Provider: provider,
		Model:    flowModel{},
	})

	run, err := runner.Run(context.Background(), orchestrator.Request{Repo: "acme/widgets", Number: 5})
	if err != nil {
		t.Fatalf("runner returned error: %v", err)
	}
	if run.State != orchestrator.StateCompleted {
		t.Fatalf("run state = %s", run.State)
	}
	if provider.summary == "" || len(provider.inline) != 1 {
		t.Fatalf("summary/inline = %q/%d", provider.summary, len(provider.inline))
	}
}

func signature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

type flowProvider struct {
	patch   provider.PatchSet
	summary string
	inline  []provider.InlineComment
}

func (f *flowProvider) GetPatchSet(context.Context, string, int) (provider.PatchSet, error) {
	return f.patch, nil
}

func (f *flowProvider) PublishSummary(_ context.Context, _ string, _ int, comment provider.SummaryComment) error {
	f.summary = comment.Body
	return nil
}

func (f *flowProvider) PublishInline(_ context.Context, _ string, _ int, _ string, comment provider.InlineComment) error {
	f.inline = append(f.inline, comment)
	return nil
}

func (f *flowProvider) UpdateStatus(context.Context, string, string, provider.ReviewStatus) error {
	return nil
}

type flowModel struct{}

func (flowModel) Review(context.Context, model.ReviewModelInput, model.ModelPolicy) (model.ReviewModelOutput, error) {
	return model.ReviewModelOutput{
		Summary:   "reviewed",
		RiskLevel: "medium",
		Findings: []model.Finding{{
			Severity:      "P1",
			Category:      "correctness",
			File:          "main.go",
			Line:          1,
			Message:       "bug",
			EvidenceLevel: "A",
			Confidence:    0.95,
		}},
	}, nil
}
