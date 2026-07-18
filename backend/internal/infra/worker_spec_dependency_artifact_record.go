package infra

import (
	"encoding/json"
	"time"
)

type workerSpecDependencyArtifactRecord struct {
	ID                   int64           `gorm:"primaryKey"`
	OrganizationID       int64           `gorm:"column:organization_id"`
	WorkerSpecSnapshotID int64           `gorm:"column:worker_spec_snapshot_id"`
	ArtifactJSON         json.RawMessage `gorm:"column:artifact_json;type:jsonb"`
	ArtifactDigest       string          `gorm:"column:artifact_digest"`
	CreatedAt            time.Time       `gorm:"column:created_at"`
}

func (workerSpecDependencyArtifactRecord) TableName() string {
	return "worker_spec_dependency_artifacts"
}
