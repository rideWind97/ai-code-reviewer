package secret

import (
	"regexp"
	"strings"
)

// patterns 是一组轻量 secret 检测规则。
//
// 这不是完整的 secret scanning 引擎，只是发布评论前的最后一道 MVP 防线。
var patterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
}

// Scan 扫描文本中疑似密钥的片段。
//
// Publisher 在真正发布评论前可以调用它，避免把模型输出中的敏感信息写回 PR。
func Scan(text string) []string {
	var matches []string
	seen := make(map[string]struct{})
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllString(text, -1) {
			if contained(matches, match) {
				continue
			}
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			matches = append(matches, match)
		}
	}
	return matches
}

// contained 判断 candidate 是否已经被更长的匹配覆盖。
//
// 例如 `api_key = "sk-..."` 和 `sk-...` 可能同时命中，只保留更完整的一条即可。
func contained(matches []string, candidate string) bool {
	for _, match := range matches {
		if strings.Contains(match, candidate) {
			return true
		}
	}
	return false
}
