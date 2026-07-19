package v1

import coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"

type createCoordinatorProjectRequest struct {
	RepositoryID         int64                      `json:"repository_id" binding:"required"`
	Name                 string                     `json:"name" binding:"required"`
	PlatformType         string                     `json:"platform_type"`
	SourceType           string                     `json:"source_type"`
	LabelFilter          []string                   `json:"label_filter"`
	ClaimPolicy          coordinatordom.ClaimPolicy `json:"claim_policy"`
	WorkerSpecSnapshotID *int64                     `json:"worker_spec_snapshot_id"`
	ScanIntervalSeconds  int                        `json:"scan_interval_seconds"`
	MaxConcurrent        int                        `json:"max_concurrent"`
}

type updateCoordinatorProjectRequest struct {
	Name                 *string   `json:"name"`
	LabelFilter          *[]string `json:"label_filter"`
	WorkerSpecSnapshotID *int64    `json:"worker_spec_snapshot_id"`
	ScanIntervalSeconds  *int      `json:"scan_interval_seconds"`
	MaxConcurrent        *int      `json:"max_concurrent"`
	Enabled              *bool     `json:"enabled"`
}
