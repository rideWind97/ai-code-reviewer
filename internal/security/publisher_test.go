package security

import (
	"context"
	"testing"

	"aicr/internal/provider"
)

type fakePublisher struct {
	summary string
	inline  []provider.InlineComment
}

func (f *fakePublisher) PublishSummary(_ context.Context, _ string, _ int, comment provider.SummaryComment) error {
	f.summary = comment.Body
	return nil
}

func (f *fakePublisher) PublishInline(_ context.Context, _ string, _ int, _ string, comment provider.InlineComment) error {
	f.inline = append(f.inline, comment)
	return nil
}

func TestSafePublisherBlocksSecretSummary(t *testing.T) {
	base := &fakePublisher{}
	safe := NewSafePublisher(base)

	err := safe.PublishSummary(context.Background(), "acme/widgets", 5, provider.SummaryComment{Body: `api_key = "sk-1234567890abcdef1234567890abcdef"`})
	if err == nil {
		t.Fatal("PublishSummary returned nil error, want secret error")
	}
	if base.summary != "" {
		t.Fatalf("base summary = %q, want empty", base.summary)
	}
}

func TestSafePublisherAllowsCleanInlineComment(t *testing.T) {
	base := &fakePublisher{}
	safe := NewSafePublisher(base)

	err := safe.PublishInline(context.Background(), "acme/widgets", 5, "abc123", provider.InlineComment{Path: "main.go", Line: 1, Body: "bug"})
	if err != nil {
		t.Fatalf("PublishInline returned error: %v", err)
	}
	if len(base.inline) != 1 {
		t.Fatalf("inline = %d, want 1", len(base.inline))
	}
}
