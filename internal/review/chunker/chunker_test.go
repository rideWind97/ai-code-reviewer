package chunker

import (
	"testing"

	"aicr/internal/provider"
)

func TestChunkPatchSetCreatesReviewUnitPerChangedFile(t *testing.T) {
	patch := provider.PatchSet{
		Provider: "github",
		Repo:     "acme/widgets",
		Number:   5,
		HeadSHA:  "abc123",
		Files: []provider.ChangedFile{
			{Path: "a.go", Status: "modified", Patch: "@@ -1 +1 @@\n-old\n+new"},
			{Path: "b.go", Status: "added", Patch: "@@ -0,0 +1 @@\n+package b"},
		},
	}

	units := ChunkPatchSet(patch)

	if len(units) != 2 {
		t.Fatalf("units = %d, want 2", len(units))
	}
	if units[0].ID != "github:acme/widgets:5:abc123:a.go" {
		t.Fatalf("first unit ID = %q", units[0].ID)
	}
	if units[1].ReviewPath != ReviewPathNewFile {
		t.Fatalf("second review path = %q, want new-file", units[1].ReviewPath)
	}
}

func TestChunkPatchSetWithHunksCreatesUnitPerDiffHunk(t *testing.T) {
	patch := provider.PatchSet{
		Provider: "github",
		Repo:     "acme/widgets",
		Number:   5,
		HeadSHA:  "abc123",
		Files: []provider.ChangedFile{{
			Path:   "main.go",
			Status: "modified",
			Patch:  "@@ -1,2 +1,2 @@\n-old\n+new\n@@ -10,2 +10,2 @@\n-before\n+after",
		}},
	}

	units := ChunkPatchSetWithOptions(patch, Options{SplitHunks: true})

	if len(units) != 2 {
		t.Fatalf("units = %d, want 2", len(units))
	}
	if units[0].Hunk.Index != 1 || units[1].Hunk.Index != 2 {
		t.Fatalf("hunk indexes = %d/%d", units[0].Hunk.Index, units[1].Hunk.Index)
	}
	if units[0].File.Patch == units[1].File.Patch {
		t.Fatal("hunk patches should be different")
	}
}
