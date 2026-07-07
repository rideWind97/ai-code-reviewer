package summary

import (
	"testing"

	"aicr/internal/provider"
)

func TestBuildModuleSummariesGroupsByTopLevelDirectory(t *testing.T) {
	summaries := BuildModuleSummaries(provider.PatchSet{
		Files: []provider.ChangedFile{
			{Path: "internal/api/handler.go", Patch: "@@ -1 +1 @@\n-old\n+new\n+extra"},
			{Path: "internal/store/store.go", Patch: "@@ -1 +1 @@\n-old\n+new"},
			{Path: "README.md", Patch: "@@ -1 +1 @@\n-old\n+new"},
		},
	})

	if len(summaries) != 2 {
		t.Fatalf("summaries = %d, want 2", len(summaries))
	}
	if summaries[0].Module != "internal" {
		t.Fatalf("first module = %q, want internal", summaries[0].Module)
	}
	if summaries[0].FileCount != 2 {
		t.Fatalf("internal file count = %d, want 2", summaries[0].FileCount)
	}
	if summaries[0].AddedLines != 3 || summaries[0].DeletedLines != 2 {
		t.Fatalf("internal added/deleted = %d/%d", summaries[0].AddedLines, summaries[0].DeletedLines)
	}
	if summaries[1].Module != "." {
		t.Fatalf("second module = %q, want root module", summaries[1].Module)
	}
}
