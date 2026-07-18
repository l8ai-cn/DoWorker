package infra

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"gorm.io/gorm"
)

func createWorkerSpecDependencyArtifact(
	db *gorm.DB,
	organizationID int64,
	snapshotID int64,
	artifactJSON []byte,
	artifactDigest string,
) error {
	document, err := workerdependency.Decode(artifactJSON)
	if err != nil {
		return fmt.Errorf("%w: invalid worker dependency artifact: %v", control.ErrCorrupt, err)
	}
	digest, err := workerdependency.Digest(document)
	if err != nil {
		return fmt.Errorf("%w: digest worker dependency artifact: %v", control.ErrCorrupt, err)
	}
	if document.OrganizationID != organizationID || digest != artifactDigest {
		return fmt.Errorf("%w: worker dependency artifact binding mismatch", control.ErrCorrupt)
	}
	record := workerSpecDependencyArtifactRecord{
		OrganizationID: organizationID, WorkerSpecSnapshotID: snapshotID,
		ArtifactJSON: artifactJSON, ArtifactDigest: artifactDigest,
	}
	if err := db.Create(&record).Error; err != nil {
		return err
	}
	return nil
}
