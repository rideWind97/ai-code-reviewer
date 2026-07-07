package store

import (
	"context"
	"encoding/json"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/chunker"
	"aicr/internal/review/orchestrator"
)

// PostgresDB 是 Store 需要的最小数据库接口。
type PostgresDB interface {
	ExecContext(ctx context.Context, query string, args ...any) error
}

// PostgresStore 是 PostgreSQL 持久化实现骨架。
type PostgresStore struct {
	db PostgresDB
}

// NewPostgres 创建 PostgresStore。
func NewPostgres(db PostgresDB) *PostgresStore {
	return &PostgresStore{db: db}
}

// SavePatchSet 保存 PatchSet JSON。
func (s *PostgresStore) SavePatchSet(ctx context.Context, runID string, patch provider.PatchSet) error {
	payload, _ := json.Marshal(patch.Files)
	return s.db.ExecContext(ctx, `INSERT INTO patch_sets (review_run_id, base_sha, head_sha, files) VALUES ($1, $2, $3, $4)`, runID, patch.BaseSHA, patch.HeadSHA, payload)
}

// SaveReviewUnits 保存 review units。
func (s *PostgresStore) SaveReviewUnits(ctx context.Context, runID string, units []chunker.ReviewUnit) error {
	for _, unit := range units {
		payload, _ := json.Marshal(unit)
		if err := s.db.ExecContext(ctx, `INSERT INTO review_units (id, review_run_id, review_path, file_path, payload) VALUES ($1, $2, $3, $4, $5)`, unit.ID, runID, unit.ReviewPath, unit.File.Path, payload); err != nil {
			return err
		}
	}
	return nil
}

// SaveFindings 保存 findings。
func (s *PostgresStore) SaveFindings(ctx context.Context, runID string, findings []model.Finding) error {
	for _, finding := range findings {
		payload, _ := json.Marshal(finding)
		if err := s.db.ExecContext(ctx, `INSERT INTO review_findings (review_run_id, fingerprint, severity, category, file_path, line, message, evidence_level, confidence, payload) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, runID, finding.Fingerprint, finding.Severity, finding.Category, finding.File, finding.Line, finding.Message, finding.EvidenceLevel, finding.Confidence, payload); err != nil {
			return err
		}
	}
	return nil
}

// UpdateRun 保存 review run 状态。
func (s *PostgresStore) UpdateRun(ctx context.Context, run orchestrator.Run) error {
	return s.db.ExecContext(ctx, `UPDATE review_runs SET state = $2, failure_reason = $3 WHERE id = $1`, run.ID, run.State, run.FailureReason)
}
