package publish

import (
	"strings"
	"testing"

	"aicr/internal/model"
	"aicr/internal/review/aggregate"
)

func TestBuildSummaryIncludesRiskAndSummaryOnlyFindings(t *testing.T) {
	body := BuildSummary(SummaryInput{
		OverallSummary: "review completed",
		RiskLevel:      "medium",
		ModuleSummaries: []ModuleSummary{{
			Module:       "internal",
			FileCount:    2,
			AddedLines:   10,
			DeletedLines: 3,
		}},
		Aggregation: aggregate.Result{
			SummaryOnly: []model.Finding{{
				Severity: "P2",
				File:     "main.go",
				Line:     12,
				Message:  "needs confirmation",
			}},
			Dropped: []model.Finding{{Message: "style nit"}},
		},
	})

	if !strings.Contains(body, "Risk: medium") {
		t.Fatalf("summary missing risk: %s", body)
	}
	if !strings.Contains(body, "needs confirmation") {
		t.Fatalf("summary missing finding: %s", body)
	}
	if !strings.Contains(body, "Filtered findings: 1") {
		t.Fatalf("summary missing filtered count: %s", body)
	}
	if !strings.Contains(body, "internal: 2 files, +10/-3") {
		t.Fatalf("summary missing module summary: %s", body)
	}
}

func TestBuildInlineCommentsUsesInlineFindingsOnly(t *testing.T) {
	comments := BuildInlineComments(aggregate.Result{
		Inline: []model.Finding{{
			File:     "main.go",
			Line:     3,
			Message:  "bug",
			Evidence: "proof",
		}},
		SummaryOnly: []model.Finding{{File: "other.go", Line: 1, Message: "maybe"}},
	})

	if len(comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments))
	}
	if comments[0].Path != "main.go" || comments[0].Line != 3 {
		t.Fatalf("comment target = %s:%d", comments[0].Path, comments[0].Line)
	}
}
