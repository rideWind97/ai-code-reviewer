package secret

import (
	"regexp"
	"strings"
)

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
}

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

func contained(matches []string, candidate string) bool {
	for _, match := range matches {
		if strings.Contains(match, candidate) {
			return true
		}
	}
	return false
}
