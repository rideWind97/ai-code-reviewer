package skill

import "strings"

// Skill 表示仓库级 review skill 的解析结果。
type Skill struct {
	Summary string   // 原始 skill 的简短摘要。
	Rules   []string // 规则列表，保持顺序。
}

// Load 从 .ai-review/SKILL.md 内容中提取规则。
//
// MVP 先把 markdown bullet 当作规则；后续可以按章节解析 architecture/tests/security 等。
func Load(content []byte) Skill {
	text := string(content)
	var rules []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			rules = append(rules, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		}
	}
	return Skill{Summary: firstHeading(text), Rules: rules}
}

// Merge 合并平台基础规则和仓库级 skill。
//
// 基础安全/证据/降噪规则必须排在前面，仓库规则只能增强，不能覆盖。
func Merge(base Skill, repo Skill) Skill {
	merged := Skill{Summary: base.Summary}
	merged.Rules = append(merged.Rules, base.Rules...)
	merged.Rules = append(merged.Rules, repo.Rules...)
	if merged.Summary == "" {
		merged.Summary = repo.Summary
	}
	return merged
}

// firstHeading 提取 markdown 第一个标题作为摘要。
func firstHeading(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return ""
}
