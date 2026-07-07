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

// WebhookOptions 是 GitHub webhook handler 的依赖。
type WebhookOptions struct {
	Secret string      // GitHub webhook secret，用来校验 X-Hub-Signature-256。
	Queue  queue.Queue // 验签和解析成功后，把 review job 写入这个队列。
}

// WebhookHandler 创建处理 GitHub webhook 的 HTTP handler。
//
// 处理顺序：
// 1. 只接受 POST。
// 2. 读取原始 body，因为 GitHub 签名校验必须使用原始字节。
// 3. 校验 X-Hub-Signature-256。
// 4. 只处理 pull_request 事件和 opened/synchronize/reopened 动作。
// 5. 用 provider/repo/pr/head/action 组成幂等 job ID，避免重复入队。
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
		// 先验签再解析 JSON，避免处理伪造请求。
		if !validSignature(opts.Secret, body, r.Header.Get("X-Hub-Signature-256")) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		// 非 PR 事件暂时不触发 review，但返回 202 表示 webhook 已被服务接收。
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

		// job.ID 是幂等键。GitHub webhook 可能重投，同一个 head sha 不应重复审查。
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

// validSignature 校验 GitHub 的 HMAC-SHA256 签名。
//
// GitHub 会发送形如 "sha256=<hex>" 的头。服务端用同一个 secret 对 body 做 HMAC，
// 再用 hmac.Equal 做常量时间比较，避免 timing attack。
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

// reviewableAction 定义哪些 PR 动作需要触发审查。
func reviewableAction(action string) bool {
	return action == "opened" || action == "synchronize" || action == "reopened"
}

// pullRequestEvent 只保留当前 MVP 需要的 GitHub webhook 字段。
//
// GitHub webhook payload 很大，没有必要把完整 JSON 都建模进来。
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
