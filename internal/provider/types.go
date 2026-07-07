package provider

import "context"

// ReviewerProvider 是 review pipeline 依赖的 provider-neutral 接口。
//
// Orchestrator 只依赖这个接口，不直接依赖 GitHub。后续要支持 GitLab 时，
// 只要写一个 GitLab adapter 实现这些方法，review 主流程就不需要大改。
type ReviewerProvider interface {
	// GetPatchSet 拉取一次 PR/MR 审查需要的 diff、文件列表和 base/head sha。
	GetPatchSet(ctx context.Context, repo string, number int) (PatchSet, error)

	// PublishSummary 发布整次 review 的总评，一般是 PR issue comment 或 MR note。
	PublishSummary(ctx context.Context, repo string, number int, comment SummaryComment) error

	// PublishInline 发布某个具体文件行上的 inline review comment。
	PublishInline(ctx context.Context, repo string, number int, headSHA string, comment InlineComment) error

	// UpdateStatus 更新 commit status/check run，用于展示 reviewing/completed/failed。
	UpdateStatus(ctx context.Context, repo, sha string, status ReviewStatus) error
}

// ChangedFile 表示 PR/MR 中的一个变更文件。
type ChangedFile struct {
	Path     string // 文件路径，例如 "internal/foo.go"。
	Status   string // 文件状态，例如 added/modified/removed。
	Patch    string // Git provider 返回的 unified diff 片段。
	Contents string // 文件内容窗口；MVP 测试中使用，后续由 repo cache/provider 填充。
}

// PatchSet 是一次审查的 diff 输入集合。
type PatchSet struct {
	Provider string        // provider 名称，例如 github/gitlab。
	Repo     string        // 仓库全名，例如 "acme/widgets"。
	Number   int           // PR/MR 编号。
	BaseSHA  string        // base commit sha。
	HeadSHA  string        // head commit sha，发布前需要用它避免评论到旧提交。
	Files    []ChangedFile // 本次变更的文件列表。
}

// InlineComment 是 provider-neutral 的行级评论。
type InlineComment struct {
	Path string // 评论所在文件。
	Line int    // 评论所在的新文件行号。
	Body string // 评论正文。
}

// SummaryComment 是 provider-neutral 的总评评论。
type SummaryComment struct {
	Body string
}

// ReviewStatus 是 provider-neutral 的审查状态。
type ReviewStatus struct {
	State       string // 状态值，例如 pending/success/failure。
	Description string // 展示给用户看的简短描述。
}
