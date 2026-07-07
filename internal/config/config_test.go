package config

import (
	"os"
	"testing"
)

func TestLoadUsesEnvironmentAndDefaults(t *testing.T) {
	t.Setenv("AICR_ADDR", ":9090")
	t.Setenv("AICR_DATABASE_URL", "postgres://user:pass@localhost/aicr")
	t.Setenv("AICR_GITHUB_WEBHOOK_SECRET", "secret")
	t.Setenv("AICR_MODEL_BASE_URL", "https://models.example.test/v1")
	t.Setenv("AICR_MODEL_NAME", "reviewer")
	t.Setenv("AICR_MODEL_API_KEY", "key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want :9090", cfg.Addr)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost/aicr" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.GitHub.WebhookSecret != "secret" {
		t.Fatalf("GitHub.WebhookSecret = %q", cfg.GitHub.WebhookSecret)
	}
	if cfg.Model.BaseURL != "https://models.example.test/v1" {
		t.Fatalf("Model.BaseURL = %q", cfg.Model.BaseURL)
	}
	if cfg.Model.Name != "reviewer" {
		t.Fatalf("Model.Name = %q", cfg.Model.Name)
	}
}

func TestLoadDefaultsAddress(t *testing.T) {
	for _, key := range os.Environ() {
		_ = key
	}
	t.Setenv("AICR_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", cfg.Addr)
	}
}
