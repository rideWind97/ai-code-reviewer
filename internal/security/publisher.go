package security

import (
	"context"
	"fmt"

	"aicr/internal/provider"
	"aicr/internal/secret"
)

// Publisher 是发布评论所需的最小接口。
type Publisher interface {
	PublishSummary(context.Context, string, int, provider.SummaryComment) error
	PublishInline(context.Context, string, int, string, provider.InlineComment) error
}

// SafePublisher 在真正发布前做 secret scanning。
type SafePublisher struct {
	next Publisher
}

// NewSafePublisher 包装一个实际 publisher。
func NewSafePublisher(next Publisher) *SafePublisher {
	return &SafePublisher{next: next}
}

// PublishSummary 发布 summary 前先扫描疑似密钥。
func (p *SafePublisher) PublishSummary(ctx context.Context, repo string, number int, comment provider.SummaryComment) error {
	if matches := secret.Scan(comment.Body); len(matches) > 0 {
		return fmt.Errorf("summary contains possible secret")
	}
	return p.next.PublishSummary(ctx, repo, number, comment)
}

// PublishInline 发布 inline 前先扫描疑似密钥。
func (p *SafePublisher) PublishInline(ctx context.Context, repo string, number int, headSHA string, comment provider.InlineComment) error {
	if matches := secret.Scan(comment.Body); len(matches) > 0 {
		return fmt.Errorf("inline comment contains possible secret")
	}
	return p.next.PublishInline(ctx, repo, number, headSHA, comment)
}
