package store

import (
	"context"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/chunker"
	"aicr/internal/review/orchestrator"
)

// Store 是持久化层接口。
//
// 当前 MVP 先定义接口和 migrations，方便上层代码面向抽象编程。
// 后续接 PostgreSQL repository 时，实现这个接口即可。
type Store interface {
	// SavePatchSet 保存本次 review 的 diff 输入。
	SavePatchSet(context.Context, string, provider.PatchSet) error

	// SaveReviewUnits 保存 chunker 生成的审查单元。
	SaveReviewUnits(context.Context, string, []chunker.ReviewUnit) error

	// SaveFindings 保存模型和聚合阶段产出的 findings。
	SaveFindings(context.Context, string, []model.Finding) error

	// UpdateRun 保存 review run 的状态变化。
	UpdateRun(context.Context, orchestrator.Run) error
}
