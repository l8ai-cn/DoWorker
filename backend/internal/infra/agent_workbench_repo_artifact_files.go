package infra

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	sessionfiledomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func insertAgentWorkbenchArtifactFiles(
	tx *gorm.DB,
	files []agentworkbench.ArtifactFile,
) error {
	for _, file := range files {
		row := sessionfiledomain.File{
			ID: file.ID, SessionID: file.SessionID,
			Filename: file.Filename, Bytes: file.Bytes,
			ContentType: file.ContentType, MinioKey: file.MinioKey,
			CreatedAt: file.CreatedAt,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			continue
		}
		var existing sessionfiledomain.File
		if err := tx.Where(
			"id = ? AND session_id = ?",
			file.ID,
			file.SessionID,
		).Take(&existing).Error; err != nil {
			return err
		}
		if existing.Filename != file.Filename ||
			existing.Bytes != file.Bytes ||
			existing.ContentType != file.ContentType {
			return fmt.Errorf("artifact file identity conflict")
		}
	}
	return nil
}
