package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"aicr/internal/queue"
	"aicr/internal/reviewjob"
)

// WebhookOptions 是 GitLab webhook handler 的依赖。
type WebhookOptions struct {
	SecretToken string      // GitLab webhook 的 Secret Token，对应请求头 X-Gitlab-Token。
	Queue       queue.Queue // 验证和解析成功后，把 review job 写入这个队列。
}

// WebhookHandler 创建处理 GitLab Merge Request webhook 的 HTTP handler。
//
// GitLab 和 GitHub 的 webhook 差异主要在两点：
// 1. GitLab 使用 X-Gitlab-Token 做共享 token 校验，而不是 HMAC 签名。
// 2. MR 事件类型是 "Merge Request Hook"，PR 编号字段叫 iid。
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
		if !validToken(opts.SecretToken, r.Header.Get("X-Gitlab-Token")) {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Gitlab-Event") != "Merge Request Hook" {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		var event mergeRequestEvent
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&event); err != nil {
			http.Error(w, "decode event", http.StatusBadRequest)
			return
		}
		if event.ObjectKind != "merge_request" || !reviewableAction(event.ObjectAttributes.Action) || event.ObjectAttributes.Draft {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		payload, err := json.Marshal(reviewjob.Request{
			Provider: "gitlab",
			Repo:     event.Project.PathWithNamespace,
			Number:   event.ObjectAttributes.IID,
		})
		if err != nil {
			http.Error(w, "encode job", http.StatusInternalServerError)
			return
		}
		job := queue.Job{
			ID:      fmt.Sprintf("gitlab:%s:%d:%s:%s", event.Project.PathWithNamespace, event.ObjectAttributes.IID, event.ObjectAttributes.LastCommit.ID, event.ObjectAttributes.Action),
			Type:    "review",
			Payload: payload,
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

// validToken 校验 GitLab webhook secret token。
//
// 如果没有配置 SecretToken，则允许请求通过，方便本地开发；生产环境应始终配置。
func validToken(secret, got string) bool {
	return secret == "" || secret == got
}

// reviewableAction 定义哪些 MR 动作会触发审查。
func reviewableAction(action string) bool {
	return action == "open" || action == "update" || action == "reopen"
}

// mergeRequestEvent 只建模当前 MVP 需要的 GitLab webhook 字段。
type mergeRequestEvent struct {
	ObjectKind string `json:"object_kind"`
	Project    struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
	ObjectAttributes struct {
		IID          int    `json:"iid"`
		Action       string `json:"action"`
		Draft        bool   `json:"draft"`
		TargetBranch string `json:"target_branch"`
		SourceBranch string `json:"source_branch"`
		LastCommit   struct {
			ID string `json:"id"`
		} `json:"last_commit"`
	} `json:"object_attributes"`
}
