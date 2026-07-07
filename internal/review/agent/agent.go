package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"aicr/internal/model"
	reviewcontext "aicr/internal/review/context"
)

type Runtime struct {
	model model.ReviewModelClient
}

func NewRuntime(client model.ReviewModelClient) *Runtime {
	return &Runtime{model: client}
}

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

func fingerprint(f model.Finding) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		f.File,
		f.Category,
		f.Message,
		string(rune(f.Line)),
	}, "\x00")))
	return hex.EncodeToString(sum[:8])
}
