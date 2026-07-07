package aggregate

import (
	"testing"

	"aicr/internal/model"
)

func TestAggregateFiltersAndDeduplicatesFindings(t *testing.T) {
	findings := []model.Finding{
		{Severity: "P1", Category: "correctness", File: "main.go", Line: 10, Message: "bug", EvidenceLevel: "A", Confidence: 0.9},
		{Severity: "P1", Category: "correctness", File: "main.go", Line: 10, Message: "bug", EvidenceLevel: "A", Confidence: 0.9},
		{Severity: "P3", Category: "style", File: "main.go", Line: 12, Message: "nit", EvidenceLevel: "D", Confidence: 0.6},
		{Severity: "P2", Category: "test", File: "main.go", Line: 14, Message: "maybe", EvidenceLevel: "C", Confidence: 0.7},
	}

	result := Aggregate(findings, Options{InlineBudget: 5})

	if len(result.Inline) != 1 {
		t.Fatalf("inline = %d, want 1", len(result.Inline))
	}
	if len(result.SummaryOnly) != 1 {
		t.Fatalf("summary only = %d, want 1", len(result.SummaryOnly))
	}
	if len(result.Dropped) != 2 {
		t.Fatalf("dropped = %d, want 2", len(result.Dropped))
	}
}

func TestAggregateHonorsInlineBudget(t *testing.T) {
	findings := []model.Finding{
		{Severity: "P2", Category: "test", File: "a.go", Line: 1, Message: "one", EvidenceLevel: "B", Confidence: 0.8},
		{Severity: "P2", Category: "test", File: "b.go", Line: 2, Message: "two", EvidenceLevel: "B", Confidence: 0.8},
	}

	result := Aggregate(findings, Options{InlineBudget: 1})

	if len(result.Inline) != 1 {
		t.Fatalf("inline = %d, want 1", len(result.Inline))
	}
	if len(result.SummaryOnly) != 1 {
		t.Fatalf("summary only = %d, want 1", len(result.SummaryOnly))
	}
}
