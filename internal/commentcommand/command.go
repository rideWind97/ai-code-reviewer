package commentcommand

import "strings"

// Mode 表示 comment command 触发哪种审查。
type Mode string

const (
	// ModeDefault 表示普通增量审查。
	ModeDefault Mode = "default"

	// ModeFull 表示全量审查。
	ModeFull Mode = "full"

	// ModeFile 表示只审查指定文件。
	ModeFile Mode = "file"
)

// Command 是解析后的 /ai-review 指令。
type Command struct {
	Mode Mode   // 审查模式。
	File string // ModeFile 时的目标文件。
}

// Parse 解析 PR/MR 评论中的 /ai-review 指令。
//
// 支持：
// - /ai-review
// - /ai-review full
// - /ai-review file path/to/file.go
func Parse(text string) (Command, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 || fields[0] != "/ai-review" {
		return Command{}, false
	}
	if len(fields) == 1 {
		return Command{Mode: ModeDefault}, true
	}
	switch fields[1] {
	case "full":
		return Command{Mode: ModeFull}, true
	case "file":
		if len(fields) < 3 {
			return Command{}, false
		}
		return Command{Mode: ModeFile, File: fields[2]}, true
	default:
		return Command{}, false
	}
}
