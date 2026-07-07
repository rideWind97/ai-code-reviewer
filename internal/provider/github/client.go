package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aicr/internal/provider"
)

// ClientOptions 是 GitHub REST API client 的配置。
type ClientOptions struct {
	BaseURL    string       // GitHub API 地址；测试时会传 httptest.Server.URL。
	Token      string       // API token；生产中后续会换成 installation access token。
	HTTPClient *http.Client // 可注入 HTTP client，方便测试 fake server。
}

// Client 实现 provider.ReviewerProvider，把统一接口映射到 GitHub REST API。
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient 创建 GitHub API client。
//
// 如果没有传 BaseURL，就默认使用 GitHub 官方 API；如果没有传 HTTPClient，
// 就创建一个带超时的标准 http.Client，避免请求永久挂住。
func NewClient(opts ClientOptions) *Client {
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   opts.Token,
		http:    client,
	}
}

// GetPatchSet 拉取 PR 元数据和变更文件列表，并转换成 provider.PatchSet。
//
// GitHub 的 PR 基本信息和 changed files 是两个 API，所以这里会发两次请求。
func (c *Client) GetPatchSet(ctx context.Context, repo string, number int) (provider.PatchSet, error) {
	var pr pullRequestResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/pulls/%d", repo, number), nil, &pr); err != nil {
		return provider.PatchSet{}, err
	}

	var files []pullFileResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/pulls/%d/files", repo, number), nil, &files); err != nil {
		return provider.PatchSet{}, err
	}

	changed := make([]provider.ChangedFile, 0, len(files))
	for _, file := range files {
		changed = append(changed, provider.ChangedFile{
			Path:   file.Filename,
			Status: file.Status,
			Patch:  file.Patch,
		})
	}
	return provider.PatchSet{
		Provider: "github",
		Repo:     repo,
		Number:   number,
		BaseSHA:  pr.Base.SHA,
		HeadSHA:  pr.Head.SHA,
		Files:    changed,
	}, nil
}

// PublishSummary 把总评写到 GitHub issue comment。
//
// GitHub PR 同时也是 issue，所以 summary comment 使用 issues comments API。
func (c *Client) PublishSummary(ctx context.Context, repo string, number int, comment provider.SummaryComment) error {
	body := map[string]string{"body": comment.Body}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number), body, nil)
}

// PublishInline 把 finding 发布为 PR review inline comment。
//
// 当前 MVP 使用 line + RIGHT side 的最小参数；后续可以增强为支持 multi-line、
// position remap 和 outdated comment 处理。
func (c *Client) PublishInline(ctx context.Context, repo string, number int, headSHA string, comment provider.InlineComment) error {
	body := map[string]any{
		"body":      comment.Body,
		"commit_id": headSHA,
		"path":      comment.Path,
		"line":      comment.Line,
		"side":      "RIGHT",
	}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/pulls/%d/comments", repo, number), body, nil)
}

// UpdateStatus 更新 commit status，展示 AI review 的运行结果。
func (c *Client) UpdateStatus(ctx context.Context, repo, sha string, status provider.ReviewStatus) error {
	body := map[string]string{
		"state":       status.State,
		"description": status.Description,
		"context":     "aicr/review",
	}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/statuses/%s", repo, sha), body, nil)
}

// doJSON 是 GitHub client 的通用请求函数。
//
// input 不为空时会编码成 JSON 请求体；output 不为空时会把响应 JSON 解码进去。
// 这样上层方法只需要关心“调用哪个 API”和“请求/响应结构是什么”。
func (c *Client) doJSON(ctx context.Context, method, path string, input any, output any) error {
	var body *bytes.Reader
	if input == nil {
		body = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("github status: %s", resp.Status)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}

// pullRequestResponse 是 GitHub PR API 响应的最小字段集合。
type pullRequestResponse struct {
	Number int `json:"number"`
	Head   struct {
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		SHA string `json:"sha"`
	} `json:"base"`
}

// pullFileResponse 是 GitHub PR files API 响应的最小字段集合。
type pullFileResponse struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Patch    string `json:"patch"`
}
