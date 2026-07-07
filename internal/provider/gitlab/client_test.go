package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"aicr/internal/provider"
)

func TestClientGetPatchSetFetchesMergeRequestChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v4/projects/acme%2Fwidgets/merge_requests/7/changes" {
			t.Fatalf("unexpected path %s", r.URL.EscapedPath())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":       7,
			"sha":       "abc123",
			"diff_refs": map[string]any{"base_sha": "def456", "head_sha": "abc123"},
			"changes": []map[string]any{{
				"new_path": "main.go",
				"old_path": "main.go",
				"new_file": false,
				"diff":     "@@ -1 +1 @@\n-old\n+new",
			}},
		})
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})

	patch, err := client.GetPatchSet(context.Background(), "acme/widgets", 7)
	if err != nil {
		t.Fatalf("GetPatchSet returned error: %v", err)
	}
	if patch.Provider != "gitlab" {
		t.Fatalf("provider = %q, want gitlab", patch.Provider)
	}
	if patch.HeadSHA != "abc123" || patch.BaseSHA != "def456" {
		t.Fatalf("shas = %s/%s", patch.HeadSHA, patch.BaseSHA)
	}
	if len(patch.Files) != 1 || patch.Files[0].Path != "main.go" {
		t.Fatalf("files = %#v", patch.Files)
	}
}

func TestClientPublishesNoteDiscussionAndStatus(t *testing.T) {
	var sawNote bool
	var sawDiscussion bool
	var sawStatus bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/api/v4/projects/acme%2Fwidgets/merge_requests/7/notes":
			sawNote = true
			if r.Method != http.MethodPost {
				t.Fatalf("note method = %s", r.Method)
			}
			assertFormValue(t, r, "body", "summary")
		case "/api/v4/projects/acme%2Fwidgets/merge_requests/7/discussions":
			sawDiscussion = true
			if r.Method != http.MethodPost {
				t.Fatalf("discussion method = %s", r.Method)
			}
			assertFormValue(t, r, "body", "inline")
			assertFormValue(t, r, "position[new_path]", "main.go")
			assertFormValue(t, r, "position[new_line]", "12")
		case "/api/v4/projects/acme%2Fwidgets/statuses/abc123":
			sawStatus = true
			if r.Method != http.MethodPost {
				t.Fatalf("status method = %s", r.Method)
			}
			assertFormValue(t, r, "state", "success")
		default:
			t.Fatalf("unexpected path %s", r.URL.EscapedPath())
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL, HTTPClient: server.Client()})
	if err := client.PublishSummary(context.Background(), "acme/widgets", 7, provider.SummaryComment{Body: "summary"}); err != nil {
		t.Fatalf("PublishSummary returned error: %v", err)
	}
	if err := client.PublishInline(context.Background(), "acme/widgets", 7, "abc123", provider.InlineComment{Path: "main.go", Line: 12, Body: "inline"}); err != nil {
		t.Fatalf("PublishInline returned error: %v", err)
	}
	if err := client.UpdateStatus(context.Background(), "acme/widgets", "abc123", provider.ReviewStatus{State: "success", Description: "done"}); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	if !sawNote || !sawDiscussion || !sawStatus {
		t.Fatalf("saw note/discussion/status = %v/%v/%v", sawNote, sawDiscussion, sawStatus)
	}
}

func assertFormValue(t *testing.T, r *http.Request, key string, want string) {
	t.Helper()
	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm returned error: %v", err)
	}
	got, err := url.QueryUnescape(r.Form.Get(key))
	if err != nil {
		t.Fatalf("QueryUnescape returned error: %v", err)
	}
	if got != want {
		t.Fatalf("form %s = %q, want %q", key, got, want)
	}
}
