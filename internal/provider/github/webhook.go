package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"aicr/internal/queue"
)

type WebhookOptions struct {
	Secret string
	Queue  queue.Queue
}

func WebhookHandler(opts WebhookOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		if !validSignature(opts.Secret, body, r.Header.Get("X-Hub-Signature-256")) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-GitHub-Event") != "pull_request" {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		var event pullRequestEvent
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&event); err != nil {
			http.Error(w, "decode event", http.StatusBadRequest)
			return
		}
		if !reviewableAction(event.Action) || event.PullRequest.Draft {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		job := queue.Job{
			ID:      fmt.Sprintf("github:%s:%d:%s:%s", event.Repository.FullName, event.PullRequest.Number, event.PullRequest.Head.SHA, event.Action),
			Type:    "review",
			Payload: body,
		}
		if opts.Queue != nil {
			if err := opts.Queue.Enqueue(context.Background(), job); err != nil {
				http.Error(w, "enqueue event", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
	})
}

func validSignature(secret string, body []byte, header string) bool {
	if secret == "" {
		return true
	}
	const prefix = "sha256="
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return false
	}
	got, err := hex.DecodeString(header[len(prefix):])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

func reviewableAction(action string) bool {
	return action == "opened" || action == "synchronize" || action == "reopened"
}

type pullRequestEvent struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest struct {
		Number int  `json:"number"`
		Draft  bool `json:"draft"`
		Head   struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
	} `json:"pull_request"`
}
