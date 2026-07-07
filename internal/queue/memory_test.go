package queue

import (
	"context"
	"testing"
)

func TestMemoryQueueDeduplicatesJobs(t *testing.T) {
	q := NewMemory()
	job := Job{ID: "github:repo:1:sha:opened", Type: "review", Payload: []byte(`{"ok":true}`)}

	if err := q.Enqueue(context.Background(), job); err != nil {
		t.Fatalf("first enqueue returned error: %v", err)
	}
	if err := q.Enqueue(context.Background(), job); err != nil {
		t.Fatalf("second enqueue returned error: %v", err)
	}

	jobs := q.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(jobs))
	}
}

func TestMemoryQueueDequeueMarksJobInFlight(t *testing.T) {
	q := NewMemory()
	_ = q.Enqueue(context.Background(), Job{ID: "job-1", Type: "review"})

	job, ok, err := q.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("Dequeue returned error: %v", err)
	}
	if !ok {
		t.Fatal("Dequeue ok = false, want true")
	}
	if job.ID != "job-1" {
		t.Fatalf("job ID = %q, want job-1", job.ID)
	}

	_, ok, err = q.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("second Dequeue returned error: %v", err)
	}
	if ok {
		t.Fatal("second Dequeue ok = true, want false")
	}
}
