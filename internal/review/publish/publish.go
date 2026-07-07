package publish

import (
	"fmt"
	"strings"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/aggregate"
)

// SummaryInput 是构造 summary comment 所需的数据。
type SummaryInput struct {
	OverallSummary string           // 所有 review unit 的总体摘要。
	RiskLevel      string           // 整体风险等级。
	Aggregation    aggregate.Result // 聚合后的 inline/summary/dropped 分桶。
}

// BuildSummary 生成 PR/MR 总评正文。
//
// Summary comment 用来承载整体结论、需要人工确认的问题，以及被过滤 finding 数量。
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

// BuildInlineComments 把可 inline 的 findings 转成 provider-neutral 评论。
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

// inlineBody 生成单条 inline comment 文本。
//
// 如果模型给了 evidence，就附在评论后面，让作者知道这条评论的依据。
func inlineBody(f model.Finding) string {
	if f.Evidence == "" {
		return f.Message
	}
	return f.Message + "\n\nEvidence: " + f.Evidence
}
