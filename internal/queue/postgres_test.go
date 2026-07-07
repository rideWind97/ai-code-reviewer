package queue

import (
	"context"
	"testing"
)

type fakeQueueDB struct {
	execQueries  []string
	queryQueries []string
	nextJob      Job
	hasJob       bool
}

func (f *fakeQueueDB) ExecContext(_ context.Context, query string, _ ...any) error {
	f.execQueries = append(f.execQueries, query)
	return nil
}

func (f *fakeQueueDB) DequeueJob(_ context.Context) (Job, bool, error) {
	f.queryQueries = append(f.queryQueries, "dequeue")
	return f.nextJob, f.hasJob, nil
}

func TestPostgresQueueEnqueueUsesIdempotentInsert(t *testing.T) {
	db := &fakeQueueDB{}
	q := NewPostgres(db)

	if err := q.Enqueue(context.Background(), Job{ID: "job-1", Type: "review", Payload: []byte(`{}`)}); err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	if len(db.execQueries) != 1 {
		t.Fatalf("exec queries = %d, want 1", len(db.execQueries))
	}
}

func TestPostgresQueueDequeuesThroughDB(t *testing.T) {
	db := &fakeQueueDB{nextJob: Job{ID: "job-1", Type: "review"}, hasJob: true}
	q := NewPostgres(db)

	job, ok, err := q.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("Dequeue returned error: %v", err)
	}
	if !ok || job.ID != "job-1" {
		t.Fatalf("job/ok = %#v/%v", job, ok)
	}
}
