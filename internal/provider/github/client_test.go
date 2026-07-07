package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aicr/internal/provider"
)

func TestClientGetPatchSetFetchesPullRequestFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widgets/pulls/5":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 5,
				"head":   map[string]any{"sha": "abc123"},
				"base":   map[string]any{"sha": "def456"},
			})
		case "/repos/acme/widgets/pulls/5/files":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"filename": "main.go",
				"status":   "modified",
				"patch":    "@@ -1 +1 @@\n-old\n+new",
			}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})

	patch, err := client.GetPatchSet(context.Background(), "acme/widgets", 5)
	if err != nil {
		t.Fatalf("GetPatchSet returned error: %v", err)
	}
	if patch.HeadSHA != "abc123" || patch.BaseSHA != "def456" {
		t.Fatalf("shas = %s/%s", patch.HeadSHA, patch.BaseSHA)
	}
	if len(patch.Files) != 1 || patch.Files[0].Path != "main.go" {
		t.Fatalf("files = %#v", patch.Files)
	}
}

func TestClientPublishesSummaryAndStatus(t *testing.T) {
	var sawComment bool
	var sawStatus bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widgets/issues/5/comments":
			sawComment = true
			if r.Method != http.MethodPost {
				t.Fatalf("comment method = %s", r.Method)
			}
		case "/repos/acme/widgets/statuses/abc123":
			sawStatus = true
			if r.Method != http.MethodPost {
				t.Fatalf("status method = %s", r.Method)
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL, HTTPClient: server.Client()})
	if err := client.PublishSummary(context.Background(), "acme/widgets", 5, provider.SummaryComment{Body: "summary"}); err != nil {
		t.Fatalf("PublishSummary returned error: %v", err)
	}
	if err := client.UpdateStatus(context.Background(), "acme/widgets", "abc123", provider.ReviewStatus{State: "success"}); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if !sawComment || !sawStatus {
		t.Fatalf("saw comment/status = %v/%v", sawComment, sawStatus)
	}
}
