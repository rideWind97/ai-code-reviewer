package commentcommand

import "testing"

func TestParseFullReviewCommand(t *testing.T) {
	cmd, ok := Parse("/ai-review full")
	if !ok {
		t.Fatal("Parse ok = false, want true")
	}
	if cmd.Mode != ModeFull {
		t.Fatalf("mode = %s, want %s", cmd.Mode, ModeFull)
	}
}

func TestParseFileReviewCommand(t *testing.T) {
	cmd, ok := Parse("/ai-review file internal/foo.go")
	if !ok {
		t.Fatal("Parse ok = false, want true")
	}
	if cmd.Mode != ModeFile || cmd.File != "internal/foo.go" {
		t.Fatalf("command = %#v", cmd)
	}
}

func TestParseIgnoresOrdinaryComment(t *testing.T) {
	_, ok := Parse("please review this")
	if ok {
		t.Fatal("Parse ok = true, want false")
	}
}
