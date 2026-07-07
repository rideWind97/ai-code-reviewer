package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"aicr/internal/provider"
)

// ClientOptions 是 GitLab REST API client 的配置。
type ClientOptions struct {
	BaseURL    string       // GitLab base URL；GitLab.com 可用默认值，自托管实例传自己的域名。
	Token      string       // GitLab OAuth、project token、group token 或 bot token。
	HTTPClient *http.Client // 测试时可注入 httptest.Server 的 client。
}

// Client 实现 provider.ReviewerProvider，把统一接口映射到 GitLab REST API。
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient 创建 GitLab API client。
//
// BaseURL 默认是 https://gitlab.com；实际 API 路径会自动加 /api/v4。
func NewClient(opts ClientOptions) *Client {
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   opts.Token,
		http:    client,
	}
}

// GetPatchSet 拉取 GitLab MR changes，并转换成 provider.PatchSet。
func (c *Client) GetPatchSet(ctx context.Context, repo string, number int) (provider.PatchSet, error) {
	var changes mergeRequestChangesResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/projects/%s/merge_requests/%d/changes", projectID(repo), number), nil, &changes); err != nil {
		return provider.PatchSet{}, err
	}

	files := make([]provider.ChangedFile, 0, len(changes.Changes))
	for _, change := range changes.Changes {
		files = append(files, provider.ChangedFile{
			Path:   change.NewPath,
			Status: change.status(),
			Patch:  change.Diff,
		})
	}
	return provider.PatchSet{
		Provider: "gitlab",
		Repo:     repo,
		Number:   number,
		BaseSHA:  changes.DiffRefs.BaseSHA,
		HeadSHA:  changes.DiffRefs.HeadSHA,
		Files:    files,
	}, nil
}

// PublishSummary 把总评发布为 GitLab MR note。
func (c *Client) PublishSummary(ctx context.Context, repo string, number int, comment provider.SummaryComment) error {
	form := url.Values{"body": {comment.Body}}
	return c.doForm(ctx, http.MethodPost, fmt.Sprintf("/projects/%s/merge_requests/%d/notes", projectID(repo), number), form, nil)
}

// PublishInline 把 finding 发布为 GitLab MR discussion。
//
// GitLab inline discussion 需要 position 字段。MVP 只填写 new_path/new_line/head_sha，
// 后续可以补 full diff refs、old_line 和更精确的 line code。
func (c *Client) PublishInline(ctx context.Context, repo string, number int, headSHA string, comment provider.InlineComment) error {
	form := url.Values{
		"body":                    {comment.Body},
		"position[position_type]": {"text"},
		"position[new_path]":      {comment.Path},
		"position[new_line]":      {strconv.Itoa(comment.Line)},
		"position[head_sha]":      {headSHA},
	}
	return c.doForm(ctx, http.MethodPost, fmt.Sprintf("/projects/%s/merge_requests/%d/discussions", projectID(repo), number), form, nil)
}

// UpdateStatus 更新 GitLab external commit status。
func (c *Client) UpdateStatus(ctx context.Context, repo, sha string, status provider.ReviewStatus) error {
	form := url.Values{
		"state":       {status.State},
		"description": {status.Description},
		"name":        {"aicr/review"},
	}
	return c.doForm(ctx, http.MethodPost, fmt.Sprintf("/projects/%s/statuses/%s", projectID(repo), sha), form, nil)
}

// doJSON 执行 GitLab JSON API 请求。
func (c *Client) doJSON(ctx context.Context, method, path string, input any, output any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), nil)
	if err != nil {
		return fmt.Errorf("create gitlab request: %w", err)
	}
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("gitlab request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gitlab status: %s", resp.Status)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode gitlab response: %w", err)
	}
	return nil
}

// doForm 执行 GitLab form-encoded API 请求。
func (c *Client) doForm(ctx context.Context, method, path string, form url.Values, output any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create gitlab request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("gitlab request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gitlab status: %s", resp.Status)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode gitlab response: %w", err)
	}
	return nil
}

// apiURL 拼出完整 GitLab API v4 URL。
func (c *Client) apiURL(path string) string {
	return c.baseURL + "/api/v4" + path
}

// authorize 设置 GitLab token 请求头。
func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}
}

// projectID 把 repo 路径转换成 GitLab API 需要的 URL encoded project id。
func projectID(repo string) string {
	return url.PathEscape(repo)
}

// mergeRequestChangesResponse 是 MR changes API 的最小响应结构。
type mergeRequestChangesResponse struct {
	IID      int    `json:"iid"`
	SHA      string `json:"sha"`
	DiffRefs struct {
		BaseSHA string `json:"base_sha"`
		HeadSHA string `json:"head_sha"`
	} `json:"diff_refs"`
	Changes []mergeRequestChange `json:"changes"`
}

// mergeRequestChange 是单个文件 diff。
type mergeRequestChange struct {
	NewPath     string `json:"new_path"`
	OldPath     string `json:"old_path"`
	Diff        string `json:"diff"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}

// status 把 GitLab change flags 转换成 provider-neutral 文件状态。
func (c mergeRequestChange) status() string {
	switch {
	case c.NewFile:
		return "added"
	case c.DeletedFile:
		return "removed"
	case c.RenamedFile:
		return "renamed"
	default:
		return "modified"
	}
}
