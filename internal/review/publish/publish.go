package publish

import (
	"fmt"
	"strings"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/aggregate"
)

type SummaryInput struct {
	OverallSummary string
	RiskLevel      string
	Aggregation    aggregate.Result
}

func BuildSummary(input SummaryInput) string {
	var b strings.Builder
	if input.OverallSummary != "" {
		b.WriteString(input.OverallSummary)
		b.WriteString("\n\n")
	}
	b.WriteString("Risk: ")
	if input.RiskLevel == "" {
		b.WriteString("low")
	} else {
		b.WriteString(input.RiskLevel)
	}
	b.WriteString("\n")

	if len(input.Aggregation.SummaryOnly) > 0 {
		b.WriteString("\nNeeds human confirmation:\n")
		for _, finding := range input.Aggregation.SummaryOnly {
			b.WriteString(fmt.Sprintf("- %s:%d %s\n", finding.File, finding.Line, finding.Message))
		}
	}
	b.WriteString(fmt.Sprintf("\nFiltered findings: %d\n", len(input.Aggregation.Dropped)))
	return b.String()
}

func BuildInlineComments(result aggregate.Result) []provider.InlineComment {
	comments := make([]provider.InlineComment, 0, len(result.Inline))
	for _, finding := range result.Inline {
		comments = append(comments, provider.InlineComment{
			Path: finding.File,
			Line: finding.Line,
			Body: inlineBody(finding),
		})
	}
	return comments
}

func inlineBody(f model.Finding) string {
	if f.Evidence == "" {
		return f.Message
	}
	return f.Message + "\n\nEvidence: " + f.Evidence
}
