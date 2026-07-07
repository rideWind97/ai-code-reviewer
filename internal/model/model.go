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

type ReviewModelClient interface {
	Review(context.Context, ReviewModelInput, ModelPolicy) (ReviewModelOutput, error)
}

type ReviewModelInput struct {
	Prompt string
}

type ModelPolicy struct {
	MaxTokens   int
	Temperature float64
}

type ReviewModelOutput struct {
	Summary   string    `json:"summary"`
	RiskLevel string    `json:"risk_level"`
	Findings  []Finding `json:"findings"`
}

type Finding struct {
	Severity      string   `json:"severity"`
	Category      string   `json:"category"`
	File          string   `json:"file"`
	Line          int      `json:"line"`
	Message       string   `json:"message"`
	Evidence      string   `json:"evidence"`
	EvidenceLevel string   `json:"evidence_level"`
	EvidenceRefs  []string `json:"evidence_refs"`
	Suggestion    string   `json:"suggestion"`
	Confidence    float64  `json:"confidence"`
	Fingerprint   string   `json:"fingerprint"`
}

type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type OpenAICompatibleClient struct {
	cfg  ClientConfig
	http *http.Client
}

func NewOpenAICompatibleClient(cfg ClientConfig) *OpenAICompatibleClient {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &OpenAICompatibleClient{cfg: cfg, http: client}
}

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

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
