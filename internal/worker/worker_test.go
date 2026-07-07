package worker

import (
	"context"
	"encoding/json"
	"testing"

	"aicr/internal/queue"
	"aicr/internal/review/orchestrator"
)

type fakeRunner struct {
	requests []ReviewRequest
}

func (f *fakeRunner) Run(ctx context.Context, req ReviewRequest) error {
	f.requests = append(f.requests, req)
	return nil
}

func TestWorkerDequeuesReviewJobAndRunsReview(t *testing.T) {
	q := queue.NewMemory()
	payload, _ := json.Marshal(ReviewRequest{Provider: "github", Repo: "acme/widgets", Number: 5})
	if err := q.Enqueue(context.Background(), queue.Job{ID: "job-1", Type: "review", Payload: payload}); err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	runner := &fakeRunner{}
	worker := New(q, runner)

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne returned error: %v", err)
	}
	if !ok {
		t.Fatal("ProcessOne ok = false, want true")
	}
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d, want 1", len(runner.requests))
	}
	if runner.requests[0].Repo != "acme/widgets" || runner.requests[0].Number != 5 {
		t.Fatalf("request = %#v", runner.requests[0])
	}
}

func TestWorkerIgnoresEmptyQueue(t *testing.T) {
	worker := New(queue.NewMemory(), &fakeRunner{})

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne returned error: %v", err)
	}
	if ok {
		t.Fatal("ProcessOne ok = true, want false")
	}
}

type fakeOrchestrator struct {
	requests []orchestrator.Request
}

func (f *fakeOrchestrator) Run(_ context.Context, req orchestrator.Request) (*orchestrator.Run, error) {
	f.requests = append(f.requests, req)
	return orchestrator.NewRun("run-1"), nil
}

func TestProviderRouterRunnerRoutesByProvider(t *testing.T) {
	github := &fakeOrchestrator{}
	gitlab := &fakeOrchestrator{}
	runner := NewProviderRouterRunner(map[string]OrchestratorRunner{
		"github": github,
		"gitlab": gitlab,
	})

	err := runner.Run(context.Background(), ReviewRequest{Provider: "gitlab", Repo: "acme/widgets", Number: 7})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(github.requests) != 0 {
		t.Fatalf("github requests = %d, want 0", len(github.requests))
	}
	if len(gitlab.requests) != 1 {
		t.Fatalf("gitlab requests = %d, want 1", len(gitlab.requests))
	}
}
