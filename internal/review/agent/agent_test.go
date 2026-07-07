package agent

import (
	"context"
	"testing"

	"aicr/internal/model"
	reviewcontext "aicr/internal/review/context"
)

type fakeModel struct {
	output model.ReviewModelOutput
}

func (f fakeModel) Review(context.Context, model.ReviewModelInput, model.ModelPolicy) (model.ReviewModelOutput, error) {
	return f.output, nil
}

func TestRuntimeReviewsContextPack(t *testing.T) {
	runtime := NewRuntime(fakeModel{output: model.ReviewModelOutput{
		Summary:   "found issue",
		RiskLevel: "medium",
		Findings: []model.Finding{{
			Severity:      "P1",
			Category:      "correctness",
			File:          "main.go",
			Line:          10,
			Message:       "bug",
			EvidenceLevel: "A",
			Confidence:    0.9,
		}},
	}})

	result, err := runtime.Review(context.Background(), reviewcontext.ContextPack{
		UnitID: "unit-1",
		Items:  []reviewcontext.ContextItem{{Source: reviewcontext.SourceDiff, Content: "+bug"}},
	})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if result.Summary != "found issue" {
		t.Fatalf("summary = %q", result.Summary)
	}
	if result.Findings[0].Fingerprint == "" {
		t.Fatal("fingerprint is empty")
	}
}
