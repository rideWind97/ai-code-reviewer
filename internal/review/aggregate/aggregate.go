package aggregate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"aicr/internal/model"
)

// Options 控制 findings 聚合策略。
type Options struct {
	InlineBudget int // 最多发布多少条 inline comment，防止刷屏。
}

// Result 是聚合后的 findings 分桶。
type Result struct {
	Inline      []model.Finding // 证据足够、能定位到行、预算允许的问题。
	SummaryOnly []model.Finding // 有价值但不适合 inline 的问题，例如证据等级 C。
	Dropped     []model.Finding // 重复、低证据或低价值问题。
}

// Aggregate 对模型 findings 做降噪处理。
//
// 规则来自架构文档：
// - A/B 证据可以 inline。
// - C 证据只进 summary。
// - D 或无法验证的内容默认丢弃。
// - 重复 fingerprint 只保留第一条。
// - inline comment 数量受预算限制。
func Aggregate(findings []model.Finding, opts Options) Result {
	if opts.InlineBudget <= 0 {
		opts.InlineBudget = 5
	}
	// 先按严重级别排序，保证预算有限时优先发布更重要的问题。
	ordered := append([]model.Finding(nil), findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return severityRank(ordered[i].Severity) < severityRank(ordered[j].Severity)
	})

	seen := make(map[string]struct{})
	var result Result
	for _, finding := range ordered {
		key := finding.Fingerprint
		if key == "" {
			// 有些模型可能没返回 fingerprint，服务端兜底生成。
			key = fingerprint(finding)
			finding.Fingerprint = key
		}
		if _, ok := seen[key]; ok {
			result.Dropped = append(result.Dropped, finding)
			continue
		}
		seen[key] = struct{}{}

		switch finding.EvidenceLevel {
		case "A", "B":
			if canInline(finding) && len(result.Inline) < opts.InlineBudget {
				result.Inline = append(result.Inline, finding)
			} else {
				result.SummaryOnly = append(result.SummaryOnly, finding)
			}
		case "C":
			result.SummaryOnly = append(result.SummaryOnly, finding)
		default:
			result.Dropped = append(result.Dropped, finding)
		}
	}
	return result
}

// canInline 判断 finding 是否满足行级评论的最低要求。
func canInline(f model.Finding) bool {
	return f.File != "" && f.Line > 0 && f.Severity != "P3"
}

// severityRank 把 P0/P1/P2/P3 映射成可排序的数字。
func severityRank(severity string) int {
	switch severity {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	default:
		return 3
	}
}

// fingerprint 为缺失 fingerprint 的 finding 生成稳定去重键。
func fingerprint(f model.Finding) string {
	normalized := strings.ToLower(strings.Join([]string{
		f.File,
		fmt.Sprint(f.Line),
		f.Category,
		strings.TrimSpace(f.Message),
	}, "\x00"))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:8])
}
