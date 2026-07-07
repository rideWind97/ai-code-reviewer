package summary

import (
	"sort"
	"strings"

	"aicr/internal/provider"
)

// ModuleSummary 是大 PR/MR 的模块级摘要。
//
// 它不替代详细审查，只用于 summary comment 中提示“哪些模块变化最大”。
type ModuleSummary struct {
	Module       string // 顶层目录；根目录文件使用 "."。
	FileCount    int    // 该模块涉及的文件数。
	AddedLines   int    // 该模块新增行数。
	DeletedLines int    // 该模块删除行数。
}

// BuildModuleSummaries 按顶层目录聚合 PatchSet。
//
// 对 large/huge PR，先看模块摘要能快速判断风险热点，比如 internal/auth 或 migrations。
func BuildModuleSummaries(patch provider.PatchSet) []ModuleSummary {
	byModule := make(map[string]*ModuleSummary)
	for _, file := range patch.Files {
		module := topLevelModule(file.Path)
		summary := byModule[module]
		if summary == nil {
			summary = &ModuleSummary{Module: module}
			byModule[module] = summary
		}
		added, deleted := countDiffLines(file.Patch)
		summary.FileCount++
		summary.AddedLines += added
		summary.DeletedLines += deleted
	}

	summaries := make([]ModuleSummary, 0, len(byModule))
	for _, summary := range byModule {
		summaries = append(summaries, *summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].FileCount == summaries[j].FileCount {
			return summaries[i].Module < summaries[j].Module
		}
		return summaries[i].FileCount > summaries[j].FileCount
	})
	return summaries
}

// topLevelModule 返回文件所属顶层目录。
func topLevelModule(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return "."
	}
	return parts[0]
}

// countDiffLines 统计 diff 中的业务增删行。
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
