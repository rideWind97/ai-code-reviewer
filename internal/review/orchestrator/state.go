package orchestrator

import "fmt"

// RunState 表示一次 review run 当前所在阶段。
//
// 使用字符串而不是 iota，是为了后续落库、打日志和排查问题时更直观。
type RunState string

const (
	StateReceived         RunState = "received"
	StateQueued           RunState = "queued"
	StateLoadingMetadata  RunState = "loading_metadata"
	StateBuildingPatchset RunState = "building_patchset"
	StateChunking         RunState = "chunking"
	StateReviewing        RunState = "reviewing"
	StateAggregating      RunState = "aggregating"
	StatePublishing       RunState = "publishing"
	StateCompleted        RunState = "completed"
	StateFailedRetryable  RunState = "failed_retryable"
	StateFailedTerminal   RunState = "failed_terminal"
	StateCanceled         RunState = "canceled"
)

// Run 是一次 PR/MR 审查的运行状态。
type Run struct {
	ID            string   // run 的业务 ID；MVP 中用 repo:number。
	State         RunState // 当前状态。
	FailureReason string   // 失败或取消原因，便于写入数据库和日志。
}

// NewRun 创建一个处于 received 状态的新 run。
func NewRun(id string) *Run {
	return &Run{ID: id, State: StateReceived}
}

// Transition 按状态机规则切换状态。
//
// 这样可以避免代码直接从 received 跳到 publishing，导致运行记录难以理解。
func (r *Run) Transition(next RunState) error {
	if !canTransition(r.State, next) {
		return fmt.Errorf("invalid transition from %s to %s", r.State, next)
	}
	r.State = next
	return nil
}

// Cancel 取消尚未完成的 run。
//
// 常见场景是 PR 有新 commit 到来，旧 head_sha 的审查结果不能再发布。
func (r *Run) Cancel(reason string) error {
	if r.State == StateCompleted {
		return fmt.Errorf("cannot cancel completed run")
	}
	r.State = StateCanceled
	r.FailureReason = reason
	return nil
}

// canTransition 定义正常状态流转规则。
//
// failed_retryable 和 failed_terminal 可以从任意正常状态进入，方便错误处理。
func canTransition(current, next RunState) bool {
	allowed := map[RunState]RunState{
		StateReceived:         StateQueued,
		StateQueued:           StateLoadingMetadata,
		StateLoadingMetadata:  StateBuildingPatchset,
		StateBuildingPatchset: StateChunking,
		StateChunking:         StateReviewing,
		StateReviewing:        StateAggregating,
		StateAggregating:      StatePublishing,
		StatePublishing:       StateCompleted,
	}
	return allowed[current] == next || next == StateFailedRetryable || next == StateFailedTerminal
}
