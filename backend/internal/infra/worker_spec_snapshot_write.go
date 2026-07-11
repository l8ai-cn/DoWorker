package infra

import (
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workerspecservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

func createWorkerSpecSnapshot(
	db *gorm.DB,
	resolved workerspecservice.ResolvedSnapshot,
) (domain.Snapshot, error) {
	spec, summary, specJSON, summaryJSON, err := decodeResolvedSnapshot(resolved)
	if err != nil {
		return domain.Snapshot{}, err
	}
	record := workerSpecSnapshotRecord{
		OrganizationID: resolved.OrganizationID(),
		Version:        resolved.Version(),
		SpecJSON:       specJSON,
		SummaryJSON:    summaryJSON,
	}
	if err := db.Create(&record).Error; err != nil {
		return domain.Snapshot{}, err
	}
	return domain.Snapshot{
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		Spec:           spec,
		Summary:        summary,
		CreatedAt:      record.CreatedAt,
	}, nil
}
