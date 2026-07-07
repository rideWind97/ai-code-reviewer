package chunker

import (
	"fmt"

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
}

// ChunkPatchSet 把整个 PatchSet 拆成 ReviewUnit 列表。
//
// 当前 MVP 按文件拆分，足够跑通闭环；架构文档中的 DAG、symbol 级拆分可以在这里演进。
func ChunkPatchSet(patch provider.PatchSet) []ReviewUnit {
	units := make([]ReviewUnit, 0, len(patch.Files))
	for _, file := range patch.Files {
		path := ReviewPathLocalChange
		if file.Status == "added" {
			path = ReviewPathNewFile
		}
		// ID 中包含 head sha，确保同一个 PR 新提交后会得到不同 review unit。
		units = append(units, ReviewUnit{
			ID:         fmt.Sprintf("%s:%s:%d:%s:%s", patch.Provider, patch.Repo, patch.Number, patch.HeadSHA, file.Path),
			ReviewPath: path,
			File:       file,
		})
	}
	return units
}
