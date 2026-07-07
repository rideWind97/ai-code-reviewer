package githubauth

import (
	"context"
	"testing"
	"time"
)

type fakeTokenSource struct {
	calls int
}

func (f *fakeTokenSource) CreateInstallationToken(context.Context, int64) (InstallationToken, error) {
	f.calls++
	return InstallationToken{Token: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func TestCacheReusesInstallationToken(t *testing.T) {
	source := &fakeTokenSource{}
	cache := NewCache(source)

	first, err := cache.Token(context.Background(), 42)
	if err != nil {
		t.Fatalf("first Token returned error: %v", err)
	}
	second, err := cache.Token(context.Background(), 42)
	if err != nil {
		t.Fatalf("second Token returned error: %v", err)
	}

	if first != "token" || second != "token" {
		t.Fatalf("tokens = %q/%q", first, second)
	}
	if source.calls != 1 {
		t.Fatalf("source calls = %d, want 1", source.calls)
	}
}
