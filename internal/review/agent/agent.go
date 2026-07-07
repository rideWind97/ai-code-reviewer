package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"aicr/internal/model"
	reviewcontext "aicr/internal/review/context"
)

// Runtime 是单个 review agent 的运行时。
//
// 它负责把 ContextPack 转成 prompt，调用模型，并补齐 finding fingerprint。
type Runtime struct {
	model model.ReviewModelClient
}

// NewRuntime 创建 agent runtime。
func NewRuntime(client model.ReviewModelClient) *Runtime {
	return &Runtime{model: client}
}

// Review 使用模型审查一个 ContextPack。
//
// 这里不直接做聚合和发布，只返回模型产出的结构化结果；这样 agent 层职责保持单一。
func (r *Runtime) Review(ctx context.Context, pack reviewcontext.ContextPack) (model.ReviewModelOutput, error) {
	output, err := r.model.Review(ctx, model.ReviewModelInput{Prompt: buildPrompt(pack)}, model.ModelPolicy{})
	if err != nil {
		return model.ReviewModelOutput{}, err
	}
	for i := range output.Findings {
		if output.Findings[i].Fingerprint == "" {
			output.Findings[i].Fingerprint = fingerprint(output.Findings[i])
		}
	}
	return output, nil
}

// buildPrompt 把结构化 ContextPack 拼成模型可读的文本 prompt。
//
// MVP 用简单文本格式；后续可以改成更严格的 XML/JSON prompt 模板。
func buildPrompt(pack reviewcontext.ContextPack) string {
	var b strings.Builder
	b.WriteString("Review this change request context. Return JSON with summary, risk_level, and findings.\n")
	b.WriteString("Unit: ")
	b.WriteString(pack.UnitID)
	b.WriteString("\n")
	for _, item := range pack.Items {
		b.WriteString("\n[")
		b.WriteString(string(item.Source))
		b.WriteString("] ")
		b.WriteString(item.Path)
		b.WriteString("\n")
		b.WriteString(item.Content)
		b.WriteString("\n")
	}
	return b.String()
}

// fingerprint 为 finding 生成稳定 ID。
//
// Aggregator 会用 fingerprint 去重，避免同一问题被多个 unit 重复评论。
func fingerprint(f model.Finding) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		f.File,
		f.Category,
		f.Message,
		string(rune(f.Line)),
	}, "\x00")))
	return hex.EncodeToString(sum[:8])
}
