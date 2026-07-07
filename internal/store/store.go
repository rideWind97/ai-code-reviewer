package store

import (
	"context"

	"aicr/internal/model"
	"aicr/internal/provider"
	"aicr/internal/review/chunker"
	"aicr/internal/review/orchestrator"
)

type Store interface {
	SavePatchSet(context.Context, provider.PatchSet) error
	SaveReviewUnits(context.Context, []chunker.ReviewUnit) error
	SaveFindings(context.Context, string, []model.Finding) error
	UpdateRun(context.Context, orchestrator.Run) error
}
