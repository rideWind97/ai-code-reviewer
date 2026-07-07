package skill

import "testing"

func TestLoadParsesReviewSkillSections(t *testing.T) {
	skill := Load([]byte(`# Review Skill

## Architecture
- no dependency from internal/provider to internal/review

## Tests
- run go test ./...
`))

	if len(skill.Rules) != 2 {
		t.Fatalf("rules = %d, want 2", len(skill.Rules))
	}
	if skill.Summary == "" {
		t.Fatal("summary is empty")
	}
}

func TestMergeKeepsBaseSafetyRules(t *testing.T) {
	base := Skill{Rules: []string{"no secrets", "evidence required"}}
	repo := Skill{Rules: []string{"run go test ./..."}}

	merged := Merge(base, repo)

	if len(merged.Rules) != 3 {
		t.Fatalf("rules = %d, want 3", len(merged.Rules))
	}
	if merged.Rules[0] != "no secrets" {
		t.Fatalf("first rule = %q", merged.Rules[0])
	}
}
