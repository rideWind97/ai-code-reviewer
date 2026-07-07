package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ReviewModelClient 是模型网关接口。
//
// Review pipeline 只依赖这个接口，因此可以接 OpenAI、DeepSeek、GLM 或私有模型网关。
type ReviewModelClient interface {
	Review(context.Context, ReviewModelInput, ModelPolicy) (ReviewModelOutput, error)
}

// ReviewModelInput 是一次模型审查调用的输入。
type ReviewModelInput struct {
	Prompt string // 已经由 agent 组装好的审查 prompt。
}

// ModelPolicy 是模型调用策略。
//
// 当前 MVP 只保留 token 和温度，后续可以扩展 reasoning effort、region、fallback 等。
type ModelPolicy struct {
	MaxTokens   int     // 最大输出 token 数；0 表示使用 provider 默认值。
	Temperature float64 // 输出随机性；review 场景通常应接近 0。
}

// ReviewModelOutput 是模型必须返回的结构化结果。
type ReviewModelOutput struct {
	Summary   string    `json:"summary"`    // 当前 review unit 的摘要。
	RiskLevel string    `json:"risk_level"` // low/medium/high/critical。
	Findings  []Finding `json:"findings"`   // 模型发现的问题列表。
}

// Finding 是一个可被聚合、过滤和发布的审查发现。
type Finding struct {
	Severity      string   `json:"severity"`       // P0/P1/P2/P3。
	Category      string   `json:"category"`       // security/correctness/performance 等。
	File          string   `json:"file"`           // 目标文件。
	Line          int      `json:"line"`           // 目标行号，必须能映射到 diff 才能 inline。
	Message       string   `json:"message"`        // 面向 PR 作者的评论内容。
	Evidence      string   `json:"evidence"`       // 为什么这是问题。
	EvidenceLevel string   `json:"evidence_level"` // A/B/C/D，决定是否能 inline。
	EvidenceRefs  []string `json:"evidence_refs"`  // 证据引用，例如 "src/foo.go:42"。
	Suggestion    string   `json:"suggestion"`     // 建议修改方向。
	Confidence    float64  `json:"confidence"`     // 置信度，0 到 1。
	Fingerprint   string   `json:"fingerprint"`    // 稳定去重 ID。
}

// ClientConfig 是 OpenAI-compatible client 的配置。
type ClientConfig struct {
	BaseURL    string       // API base URL，不包含 /chat/completions。
	APIKey     string       // Bearer token。
	Model      string       // 模型名。
	HTTPClient *http.Client // 测试时可注入 fake client。
}

// OpenAICompatibleClient 调用兼容 OpenAI Chat Completions 的模型服务。
type OpenAICompatibleClient struct {
	cfg  ClientConfig
	http *http.Client
}

// NewOpenAICompatibleClient 创建模型 client，并设置默认 HTTP timeout。
func NewOpenAICompatibleClient(cfg ClientConfig) *OpenAICompatibleClient {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &OpenAICompatibleClient{cfg: cfg, http: client}
}

// Review 调用 /chat/completions，并把 message.content 当成 ReviewModelOutput JSON 解析。
//
// 这要求上游 prompt 明确要求模型输出 JSON；如果模型输出不是合法 JSON，会返回错误，
// 由上层决定是否 repair 或降级为 summary-only。
func (c *OpenAICompatibleClient) Review(ctx context.Context, input ReviewModelInput, policy ModelPolicy) (ReviewModelOutput, error) {
	body := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{{
			Role:    "user",
			Content: input.Prompt,
		}},
		Temperature: policy.Temperature,
		MaxTokens:   policy.MaxTokens,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return ReviewModelOutput{}, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BaseURL, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return ReviewModelOutput{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		// OpenAI-compatible 服务通常使用 Bearer token。
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return ReviewModelOutput{}, fmt.Errorf("call model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ReviewModelOutput{}, fmt.Errorf("model status: %s", resp.Status)
	}

	var chat chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
		return ReviewModelOutput{}, fmt.Errorf("decode response: %w", err)
	}
	if len(chat.Choices) == 0 {
		return ReviewModelOutput{}, fmt.Errorf("decode response: missing choices")
	}

	var output ReviewModelOutput
	if err := json.Unmarshal([]byte(chat.Choices[0].Message.Content), &output); err != nil {
		return ReviewModelOutput{}, fmt.Errorf("decode model output: %w", err)
	}
	return output, nil
}

// chatRequest 是 Chat Completions API 的请求体。
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// chatMessage 是 Chat Completions API 的一条消息。
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse 是 Chat Completions API 的最小响应结构。
type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
