package context

import (
	"strings"

	"aicr/internal/review/chunker"
)

type Source string

const (
	SourceDiff             Source = "diff"
	SourceFileWindow       Source = "file-window"
	SourceExistingComments Source = "existing-comments"
	SourceProjectSummary   Source = "project-summary"
)

type ContextPack struct {
	UnitID string
	Items  []ContextItem
}

type ContextItem struct {
	Source    Source
	Path      string
	LineRange string
	Reason    string
	Content   string
}

type BuildOptions struct {
	ExistingComments []string
	ProjectSummary   string
}

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
