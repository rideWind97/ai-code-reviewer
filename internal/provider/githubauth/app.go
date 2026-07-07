package githubauth

import (
	"context"
	"sync"
	"time"
)

// InstallationToken 是 GitHub App installation access token。
type InstallationToken struct {
	Token     string    // 实际 token。
	ExpiresAt time.Time // 过期时间。
}

// TokenSource 负责向 GitHub 换取 installation token。
//
// 真实实现需要用 App private key 生成 JWT，再调用 GitHub API。
type TokenSource interface {
	CreateInstallationToken(context.Context, int64) (InstallationToken, error)
}

// Cache 缓存 installation token，避免每次 API 调用都向 GitHub 换 token。
type Cache struct {
	source TokenSource
	mu     sync.Mutex
	tokens map[int64]InstallationToken
}

// NewCache 创建 token cache。
func NewCache(source TokenSource) *Cache {
	return &Cache{source: source, tokens: make(map[int64]InstallationToken)}
}

// Token 返回 installation token。
//
// 如果缓存 token 距离过期还超过 1 分钟，就直接复用；否则重新换取。
func (c *Cache) Token(ctx context.Context, installationID int64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if token, ok := c.tokens[installationID]; ok && time.Until(token.ExpiresAt) > time.Minute {
		return token.Token, nil
	}
	token, err := c.source.CreateInstallationToken(ctx, installationID)
	if err != nil {
		return "", err
	}
	c.tokens[installationID] = token
	return token.Token, nil
}
