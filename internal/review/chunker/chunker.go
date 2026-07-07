package chunker

import (
	"fmt"

	"aicr/internal/provider"
)

type ReviewPath string

const (
	ReviewPathLocalChange ReviewPath = "local-change"
	ReviewPathNewFile     ReviewPath = "new-file"
)

type ReviewUnit struct {
	ID         string
	ReviewPath ReviewPath
	File       provider.ChangedFile
}

func ChunkPatchSet(patch provider.PatchSet) []ReviewUnit {
	units := make([]ReviewUnit, 0, len(patch.Files))
	for _, file := range patch.Files {
		path := ReviewPathLocalChange
		if file.Status == "added" {
			path = ReviewPathNewFile
		}
		units = append(units, ReviewUnit{
			ID:         fmt.Sprintf("%s:%s:%d:%s:%s", patch.Provider, patch.Repo, patch.Number, patch.HeadSHA, file.Path),
			ReviewPath: path,
			File:       file,
		})
	}
	return units
}
