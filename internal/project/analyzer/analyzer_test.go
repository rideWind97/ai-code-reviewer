package analyzer

import "testing"

func TestAnalyzeBuildsProjectAnalysisSummary(t *testing.T) {
	result := Analyze(Snapshot{
		Files: []File{
			{Path: "go.mod", Content: "module aicr\n\ngo 1.23\n"},
			{Path: "cmd/aicr-server/main.go", Content: "package main\nfunc main() {}\n"},
			{Path: "internal/review/runner.go", Content: "package review\n"},
			{Path: "internal/review/runner_test.go", Content: "package review\n"},
		},
	})

	if result.Language != "Go" {
		t.Fatalf("language = %q, want Go", result.Language)
	}
	if len(result.EntryPoints) != 1 || result.EntryPoints[0] != "cmd/aicr-server/main.go" {
		t.Fatalf("entry points = %#v", result.EntryPoints)
	}
	if result.TestFiles != 1 {
		t.Fatalf("test files = %d, want 1", result.TestFiles)
	}
	if result.RenderMarkdown() == "" {
		t.Fatal("RenderMarkdown returned empty string")
	}
}
