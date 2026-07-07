package orchestrator

import "fmt"

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

type Run struct {
	ID            string
	State         RunState
	FailureReason string
}

func NewRun(id string) *Run {
	return &Run{ID: id, State: StateReceived}
}

func (r *Run) Transition(next RunState) error {
	if !canTransition(r.State, next) {
		return fmt.Errorf("invalid transition from %s to %s", r.State, next)
	}
	r.State = next
	return nil
}

func (r *Run) Cancel(reason string) error {
	if r.State == StateCompleted {
		return fmt.Errorf("cannot cancel completed run")
	}
	r.State = StateCanceled
	r.FailureReason = reason
	return nil
}

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
