package store

import (
	"context"
	"testing"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/orchestrator"
)

type fakeStoreDB struct {
	queries []string
}

func (f *fakeStoreDB) ExecContext(_ context.Context, query string, _ ...any) error {
	f.queries = append(f.queries, query)
	return nil
}

func TestPostgresStoreExecutesSaveOperations(t *testing.T) {
	db := &fakeStoreDB{}
	store := NewPostgres(db)

	if err := store.SavePatchSet(context.Background(), "run-1", provider.PatchSet{Repo: "acme/widgets"}); err != nil {
		t.Fatalf("SavePatchSet returned error: %v", err)
	}
	if err := store.SaveFindings(context.Background(), "run-1", []model.Finding{{Message: "bug"}}); err != nil {
		t.Fatalf("SaveFindings returned error: %v", err)
	}
	if err := store.UpdateRun(context.Background(), orchestrator.Run{ID: "run-1", State: orchestrator.StateCompleted}); err != nil {
		t.Fatalf("UpdateRun returned error: %v", err)
	}

	if len(db.queries) != 3 {
		t.Fatalf("queries = %d, want 3", len(db.queries))
	}
}
