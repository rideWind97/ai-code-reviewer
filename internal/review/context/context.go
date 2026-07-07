package context

import (
	"strings"

	"aicr/internal/review/chunker"
)

// Source 标记 ContextItem 的来源。
//
// 模型输出 finding 时需要引用证据；source 能帮助我们解释“证据来自哪里”。
type Source string

const (
	// SourceDiff 表示 Git diff hunk。
	SourceDiff Source = "diff"
	// SourceFileWindow 表示变更文件的源码上下文窗口。
	SourceFileWindow Source = "file-window"
	// SourceExistingComments 表示 PR/MR 上已有评论，用于避免重复评论。
	SourceExistingComments Source = "existing-comments"
	// SourceProjectSummary 表示项目分析文件或 README 摘要。
	SourceProjectSummary Source = "project-summary"
)

// ContextPack 是发送给 agent/model 的证据包。
type ContextPack struct {
	UnitID string        // 对应的 ReviewUnit ID。
	Items  []ContextItem // 多条上下文证据。
}

// ContextItem 是 ContextPack 中的一条证据。
type ContextItem struct {
	Source    Source // 证据来源，例如 diff/file-window。
	Path      string // 相关文件路径。
	LineRange string // 行号范围；MVP 暂未填充，后续可用于 inline 定位。
	Reason    string // 为什么这条上下文对审查有帮助。
	Content   string // 具体上下文内容。
}

// BuildOptions 是构建 ContextPack 时可选的额外上下文。
type BuildOptions struct {
	ExistingComments []string // 已有评论，用于降噪和去重。
	ProjectSummary   string   // 项目级摘要，用于补充架构/复用信息。
}

// BuildContextPack 为单个 ReviewUnit 构建 ContextPack。
//
// MVP 先放入 diff 和文件窗口两类核心证据，再按需附加历史评论和项目摘要。
func BuildContextPack(unit chunker.ReviewUnit, opts BuildOptions) ContextPack {
	items := []ContextItem{
		{
			Source:  SourceDiff,
			Path:    unit.File.Path,
			Reason:  "changed diff for review unit",
			Content: unit.File.Patch,
		},
		{
			Source:  SourceFileWindow,
			Path:    unit.File.Path,
			Reason:  "local file context around changed code",
			Content: unit.File.Contents,
		},
	}
	if len(opts.ExistingComments) > 0 {
		items = append(items, ContextItem{
			Source:  SourceExistingComments,
			Reason:  "avoid duplicate comments",
			Content: strings.Join(opts.ExistingComments, "\n"),
		})
	}
	if opts.ProjectSummary != "" {
		items = append(items, ContextItem{
			Source:  SourceProjectSummary,
			Reason:  "project-level architecture and reuse context",
			Content: opts.ProjectSummary,
		})
	}
	return ContextPack{UnitID: unit.ID, Items: items}
}
