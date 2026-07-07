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

type ClientOptions struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

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

func (c *Client) PublishSummary(ctx context.Context, repo string, number int, comment provider.SummaryComment) error {
	body := map[string]string{"body": comment.Body}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number), body, nil)
}

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

func (c *Client) UpdateStatus(ctx context.Context, repo, sha string, status provider.ReviewStatus) error {
	body := map[string]string{
		"state":       status.State,
		"description": status.Description,
		"context":     "aicr/review",
	}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/statuses/%s", repo, sha), body, nil)
}

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

type pullRequestResponse struct {
	Number int `json:"number"`
	Head   struct {
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		SHA string `json:"sha"`
	} `json:"base"`
}

type pullFileResponse struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Patch    string `json:"patch"`
}
