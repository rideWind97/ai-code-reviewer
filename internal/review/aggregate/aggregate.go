package aggregate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"aicr/internal/model"
)

type Options struct {
	InlineBudget int
}

type Result struct {
	Inline      []model.Finding
	SummaryOnly []model.Finding
	Dropped     []model.Finding
}

func Aggregate(findings []model.Finding, opts Options) Result {
	if opts.InlineBudget <= 0 {
		opts.InlineBudget = 5
	}
	ordered := append([]model.Finding(nil), findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return severityRank(ordered[i].Severity) < severityRank(ordered[j].Severity)
	})

	seen := make(map[string]struct{})
	var result Result
	for _, finding := range ordered {
		key := finding.Fingerprint
		if key == "" {
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

func canInline(f model.Finding) bool {
	return f.File != "" && f.Line > 0 && f.Severity != "P3"
}

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
