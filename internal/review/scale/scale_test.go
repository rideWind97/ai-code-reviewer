package scale

import (
	"testing"

	"aicr/internal/provider"
)

func TestClassifySmallPatchSet(t *testing.T) {
	result := Classify(provider.PatchSet{
		Files: []provider.ChangedFile{{
			Path:  "main.go",
			Patch: "@@ -1 +1 @@\n-old\n+new",
		}},
	})

	if result.Level != LevelSmall {
		t.Fatalf("level = %s, want %s", result.Level, LevelSmall)
	}
	if result.InlineBudget != 5 {
		t.Fatalf("inline budget = %d, want 5", result.InlineBudget)
	}
	if result.MaxConcurrency != 4 {
		t.Fatalf("max concurrency = %d, want 4", result.MaxConcurrency)
	}
}

func TestClassifyLargePatchSet(t *testing.T) {
	patch := provider.PatchSet{}
	for i := 0; i < 25; i++ {
		patch.Files = append(patch.Files, provider.ChangedFile{
			Path:  "pkg/file.go",
			Patch: repeatAddedLines(20),
		})
	}

	result := Classify(patch)

	if result.Level != LevelLarge {
		t.Fatalf("level = %s, want %s", result.Level, LevelLarge)
	}
	if result.InlineBudget != 3 {
		t.Fatalf("inline budget = %d, want 3", result.InlineBudget)
	}
	if result.MaxConcurrency != 8 {
		t.Fatalf("max concurrency = %d, want 8", result.MaxConcurrency)
	}
}

func TestClassifyHugePatchSetDisablesInlineComments(t *testing.T) {
	patch := provider.PatchSet{}
	for i := 0; i < 90; i++ {
		patch.Files = append(patch.Files, provider.ChangedFile{
			Path:  "pkg/file.go",
			Patch: repeatAddedLines(50),
		})
	}

	result := Classify(patch)

	if result.Level != LevelHuge {
		t.Fatalf("level = %s, want %s", result.Level, LevelHuge)
	}
	if result.InlineBudget != 0 {
		t.Fatalf("inline budget = %d, want 0", result.InlineBudget)
	}
	if result.PublishInline {
		t.Fatal("PublishInline = true, want false")
	}
}

func repeatAddedLines(n int) string {
	patch := "@@ -1,0 +1," + string(rune('0'+n%10)) + " @@\n"
	for i := 0; i < n; i++ {
		patch += "+line\n"
	}
	return patch
}
