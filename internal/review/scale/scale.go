package scale

import (
	"strings"

	"aicr/internal/provider"
)

// Level 表示 PR/MR 的规模等级。
//
// 规模等级会影响审查策略：越大越需要摘要、分片和降噪。
type Level string

const (
	// LevelSmall 表示普通变更，可以按默认策略发布 inline comments。
	LevelSmall Level = "small"

	// LevelLarge 表示较大变更，需要更细粒度分片，并降低 inline 评论预算。
	LevelLarge Level = "large"

	// LevelHuge 表示超大变更，默认不刷 inline，只发布 summary 和风险热点。
	LevelHuge Level = "huge"
)

// Result 是规模分级后的策略结果。
type Result struct {
	Level          Level // small/large/huge。
	FileCount      int   // 变更文件数。
	AddedLines     int   // diff 中新增行数。
	DeletedLines   int   // diff 中删除行数。
	InlineBudget   int   // 本次 review 最多发布多少条 inline comment。
	MaxConcurrency int   // 建议最大并发 agent 数。
	PublishInline  bool  // 是否允许发布 inline comments。
}

// Classify 根据文件数和 diff 行数判断 PR/MR 规模。
//
// 阈值先用保守常量实现，后续可以改成仓库配置或组织策略。
func Classify(patch provider.PatchSet) Result {
	fileCount := len(patch.Files)
	var added int
	var deleted int
	for _, file := range patch.Files {
		a, d := countDiffLines(file.Patch)
		added += a
		deleted += d
	}

	result := Result{
		Level:          LevelSmall,
		FileCount:      fileCount,
		AddedLines:     added,
		DeletedLines:   deleted,
		InlineBudget:   5,
		MaxConcurrency: 4,
		PublishInline:  true,
	}

	totalChanged := added + deleted
	switch {
	case fileCount >= 80 || totalChanged >= 3000:
		result.Level = LevelHuge
		result.InlineBudget = 0
		result.MaxConcurrency = 12
		result.PublishInline = false
	case fileCount >= 20 || totalChanged >= 500:
		result.Level = LevelLarge
		result.InlineBudget = 3
		result.MaxConcurrency = 8
	}
	return result
}

// countDiffLines 统计 unified diff 中真实的增删行。
//
// 注意：diff header 中的 "+++" 和 "---" 不是业务代码变更，需要排除。
func countDiffLines(patch string) (added int, deleted int) {
	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			deleted++
		}
	}
	return added, deleted
}
