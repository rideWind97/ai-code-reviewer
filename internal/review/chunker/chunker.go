package chunker

import (
	"fmt"
	"strings"

	"aicr/internal/provider"
)

// ReviewPath 表示一个 review unit 应该走哪类审查路径。
//
// 不同路径的重点不同：新增文件要看职责和导出 API，局部修改更关注上下文行为变化。
type ReviewPath string

const (
	// ReviewPathLocalChange 表示已有文件中的局部变更。
	ReviewPathLocalChange ReviewPath = "local-change"

	// ReviewPathNewFile 表示新增文件。
	ReviewPathNewFile ReviewPath = "new-file"
)

// ReviewUnit 是交给单个 agent 审查的最小单元。
type ReviewUnit struct {
	ID         string               // 稳定 ID，用于日志、存储和追踪。
	ReviewPath ReviewPath           // 审查路径，决定 agent 关注点。
	File       provider.ChangedFile // 当前 MVP 一个 unit 对应一个 changed file。
	Hunk       Hunk                 // 当按 hunk 拆分时，记录当前 unit 对应的 diff hunk。
}

// Hunk 表示 unified diff 中的一个 @@ 片段。
type Hunk struct {
	Index  int    // hunk 在文件内的序号，从 1 开始。
	Header string // hunk header，例如 "@@ -1,2 +1,2 @@"。
	Patch  string // 当前 hunk 的完整 diff 内容。
}

// Options 控制 PatchSet 拆分方式。
type Options struct {
	SplitHunks bool // true 时按 diff hunk 拆分；false 时按文件拆分。
}

// ChunkPatchSet 把整个 PatchSet 拆成 ReviewUnit 列表。
//
// 当前 MVP 按文件拆分，足够跑通闭环；架构文档中的 DAG、symbol 级拆分可以在这里演进。
func ChunkPatchSet(patch provider.PatchSet) []ReviewUnit {
	return ChunkPatchSetWithOptions(patch, Options{})
}

// ChunkPatchSetWithOptions 按配置拆分 PatchSet。
//
// 大 PR/MR 可以打开 SplitHunks，把一个大文件拆成多个更小 review unit，
// 这样 agent 的上下文更聚焦，也更容易并行调度。
func ChunkPatchSetWithOptions(patch provider.PatchSet, opts Options) []ReviewUnit {
	units := make([]ReviewUnit, 0, len(patch.Files))
	for _, file := range patch.Files {
		path := ReviewPathLocalChange
		if file.Status == "added" {
			path = ReviewPathNewFile
		}
		if !opts.SplitHunks {
			// ID 中包含 head sha，确保同一个 PR 新提交后会得到不同 review unit。
			units = append(units, ReviewUnit{
				ID:         fmt.Sprintf("%s:%s:%d:%s:%s", patch.Provider, patch.Repo, patch.Number, patch.HeadSHA, file.Path),
				ReviewPath: path,
				File:       file,
			})
			continue
		}

		hunks := splitHunks(file.Patch)
		if len(hunks) == 0 {
			units = append(units, ReviewUnit{
				ID:         fmt.Sprintf("%s:%s:%d:%s:%s", patch.Provider, patch.Repo, patch.Number, patch.HeadSHA, file.Path),
				ReviewPath: path,
				File:       file,
			})
			continue
		}
		for _, hunk := range hunks {
			hunkFile := file
			hunkFile.Patch = hunk.Patch
			units = append(units, ReviewUnit{
				ID:         fmt.Sprintf("%s:%s:%d:%s:%s:hunk:%d", patch.Provider, patch.Repo, patch.Number, patch.HeadSHA, file.Path, hunk.Index),
				ReviewPath: path,
				File:       hunkFile,
				Hunk:       hunk,
			})
		}
	}
	return units
}

// splitHunks 把 unified diff 按 @@ header 拆成多个 hunk。
func splitHunks(patch string) []Hunk {
	var hunks []Hunk
	var current []string
	var header string
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "@@ ") {
			if len(current) > 0 {
				hunks = append(hunks, Hunk{Index: len(hunks) + 1, Header: header, Patch: strings.Join(current, "\n")})
			}
			header = line
			current = []string{line}
			continue
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		hunks = append(hunks, Hunk{Index: len(hunks) + 1, Header: header, Patch: strings.Join(current, "\n")})
	}
	return hunks
}
