package gitlab

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"aicr/internal/queue"
)

func TestWebhookRejectsInvalidToken(t *testing.T) {
	q := queue.NewMemory()
	handler := WebhookHandler(WebhookOptions{
		SecretToken: "top-secret",
		Queue:       q,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/gitlab", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Gitlab-Event", "Merge Request Hook")
	req.Header.Set("X-Gitlab-Token", "bad")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if len(q.Jobs()) != 0 {
		t.Fatalf("queued jobs = %d, want 0", len(q.Jobs()))
	}
}

func TestWebhookEnqueuesMergeRequestOnce(t *testing.T) {
	q := queue.NewMemory()
	handler := WebhookHandler(WebhookOptions{
		SecretToken: "top-secret",
		Queue:       q,
	})
	body := []byte(`{
		"object_kind": "merge_request",
		"project": {"path_with_namespace": "acme/widgets"},
		"object_attributes": {
			"iid": 7,
			"action": "open",
			"draft": false,
			"last_commit": {"id": "abc123"},
			"target_branch": "main",
			"source_branch": "feature"
		}
	}`)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/gitlab", bytes.NewReader(body))
		req.Header.Set("X-Gitlab-Event", "Merge Request Hook")
		req.Header.Set("X-Gitlab-Token", "top-secret")
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
	if jobs[0].ID != "gitlab:acme/widgets:7:abc123:open" {
		t.Fatalf("job id = %q", jobs[0].ID)
	}
}

func TestWebhookSkipsDraftMergeRequest(t *testing.T) {
	q := queue.NewMemory()
	handler := WebhookHandler(WebhookOptions{
		SecretToken: "top-secret",
		Queue:       q,
	})
	body := []byte(`{
		"object_kind": "merge_request",
		"project": {"path_with_namespace": "acme/widgets"},
		"object_attributes": {
			"iid": 7,
			"action": "open",
			"draft": true,
			"last_commit": {"id": "abc123"}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/gitlab", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Merge Request Hook")
	req.Header.Set("X-Gitlab-Token", "top-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if len(q.Jobs()) != 0 {
		t.Fatalf("queued jobs = %d, want 0", len(q.Jobs()))
	}
}
