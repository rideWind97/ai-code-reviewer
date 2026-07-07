package provider

import "context"

type ReviewerProvider interface {
	GetPatchSet(ctx context.Context, repo string, number int) (PatchSet, error)
	PublishSummary(ctx context.Context, repo string, number int, comment SummaryComment) error
	PublishInline(ctx context.Context, repo string, number int, headSHA string, comment InlineComment) error
	UpdateStatus(ctx context.Context, repo, sha string, status ReviewStatus) error
}

type ChangedFile struct {
	Path     string
	Status   string
	Patch    string
	Contents string
}

type PatchSet struct {
	Provider string
	Repo     string
	Number   int
	BaseSHA  string
	HeadSHA  string
	Files    []ChangedFile
}

type InlineComment struct {
	Path string
	Line int
	Body string
}

type SummaryComment struct {
	Body string
}

type ReviewStatus struct {
	State       string
	Description string
}
