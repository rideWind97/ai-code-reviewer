package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"aicr/internal/queue"
)

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	q := queue.NewMemory()
	handler := WebhookHandler(WebhookOptions{
		Secret: "top-secret",
		Queue:  q,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewBufferString(`{}`))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=bad")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if len(q.Jobs()) != 0 {
		t.Fatalf("queued jobs = %d, want 0", len(q.Jobs()))
	}
}

func TestWebhookEnqueuesPullRequestOnce(t *testing.T) {
	q := queue.NewMemory()
	handler := WebhookHandler(WebhookOptions{
		Secret: "top-secret",
		Queue:  q,
	})
	body := []byte(`{
		"action": "opened",
		"installation": {"id": 12},
		"repository": {"id": 34, "full_name": "acme/widgets"},
		"pull_request": {
			"number": 5,
			"draft": false,
			"head": {"sha": "abc123"},
			"base": {"sha": "def456"}
		}
	}`)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
		req.Header.Set("X-GitHub-Event", "pull_request")
		req.Header.Set("X-Hub-Signature-256", sign("top-secret", body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
		}
	}

	jobs := q.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("queued jobs = %d, want 1", len(jobs))
	}
	if jobs[0].ID != "github:acme/widgets:5:abc123:opened" {
		t.Fatalf("job id = %q", jobs[0].ID)
	}
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
