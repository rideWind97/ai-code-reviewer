package secret

import "testing"

func TestScanFindsLikelySecrets(t *testing.T) {
	matches := Scan("api_key = \"sk-1234567890abcdef1234567890abcdef\"")
	if len(matches) != 1 {
		t.Fatalf("matches = %d, want 1", len(matches))
	}
}

func TestScanIgnoresOrdinaryText(t *testing.T) {
	matches := Scan("review completed without high value findings")
	if len(matches) != 0 {
		t.Fatalf("matches = %d, want 0", len(matches))
	}
}
