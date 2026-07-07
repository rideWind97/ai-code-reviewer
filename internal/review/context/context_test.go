package context

import (
	"testing"

	"aicr/internal/provider"
	"aicr/internal/review/chunker"
)

func TestBuildContextPackIncludesDiffAndFileWindow(t *testing.T) {
	file := provider.ChangedFile{
		Path:     "main.go",
		Status:   "modified",
		Patch:    "@@ -1 +1 @@\n-package old\n+package main",
		Contents: "package main\n\nfunc main() {}\n",
	}
	unit := chunker.ReviewUnit{
		ID:         "unit-1",
		File:       file,
		ReviewPath: chunker.ReviewPathLocalChange,
	}

	pack := BuildContextPack(unit, BuildOptions{
		ExistingComments: []string{"previous comment"},
		ProjectSummary:   "Go service",
	})

	if pack.UnitID != "unit-1" {
		t.Fatalf("UnitID = %q", pack.UnitID)
	}
	if len(pack.Items) != 4 {
		t.Fatalf("items = %d, want 4", len(pack.Items))
	}
	if pack.Items[0].Source != SourceDiff {
		t.Fatalf("first source = %q, want diff", pack.Items[0].Source)
	}
	if pack.Items[1].Source != SourceFileWindow {
		t.Fatalf("second source = %q, want file-window", pack.Items[1].Source)
	}
}
