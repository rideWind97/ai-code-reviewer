package orchestrator

import "testing"

func TestRunStateTransitions(t *testing.T) {
	run := NewRun("run-1")

	steps := []RunState{
		StateQueued,
		StateLoadingMetadata,
		StateBuildingPatchset,
		StateChunking,
		StateReviewing,
		StateAggregating,
		StatePublishing,
		StateCompleted,
	}

	for _, state := range steps {
		if err := run.Transition(state); err != nil {
			t.Fatalf("Transition(%s) returned error: %v", state, err)
		}
	}

	if run.State != StateCompleted {
		t.Fatalf("state = %s, want %s", run.State, StateCompleted)
	}
}

func TestRunRejectsInvalidTransition(t *testing.T) {
	run := NewRun("run-1")

	if err := run.Transition(StatePublishing); err == nil {
		t.Fatal("Transition to publishing from received succeeded, want error")
	}
}

func TestRunCanBeCanceledBeforeCompletion(t *testing.T) {
	run := NewRun("run-1")
	if err := run.Transition(StateQueued); err != nil {
		t.Fatalf("Transition queued returned error: %v", err)
	}

	if err := run.Cancel("new head sha"); err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}

	if run.State != StateCanceled {
		t.Fatalf("state = %s, want %s", run.State, StateCanceled)
	}
	if run.FailureReason != "new head sha" {
		t.Fatalf("failure reason = %q", run.FailureReason)
	}
}
