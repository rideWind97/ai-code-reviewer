package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleClientParsesStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"summary\":\"ok\",\"risk_level\":\"low\",\"findings\":[]}"
				}
			}]
		}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "reviewer",
	})

	out, err := client.Review(context.Background(), ReviewModelInput{Prompt: "review this"}, ModelPolicy{})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if out.Summary != "ok" {
		t.Fatalf("summary = %q, want ok", out.Summary)
	}
}

func TestOpenAICompatibleClientRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"not-json"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(ClientConfig{BaseURL: server.URL, Model: "reviewer"})

	_, err := client.Review(context.Background(), ReviewModelInput{Prompt: "review this"}, ModelPolicy{})
	if err == nil {
		t.Fatal("Review returned nil error, want invalid JSON error")
	}
	if !strings.Contains(err.Error(), "decode model output") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRepairJSONExtractsObjectFromMarkdownFence(t *testing.T) {
	out, err := DecodeReviewOutput("```json\n{\"summary\":\"ok\",\"risk_level\":\"low\",\"findings\":[]}\n```")
	if err != nil {
		t.Fatalf("DecodeReviewOutput returned error: %v", err)
	}
	if out.Summary != "ok" {
		t.Fatalf("summary = %q, want ok", out.Summary)
	}
}
