package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"gorm.io/gorm"
)

type workerSpecDependencyArtifactRepo struct {
	db *gorm.DB
}

func NewWorkerSpecDependencyArtifactRepository(
	db *gorm.DB,
) *workerSpecDependencyArtifactRepo {
	return &workerSpecDependencyArtifactRepo{db: db}
}

func (repo *workerSpecDependencyArtifactRepo) Create(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
	artifactJSON []byte,
	artifactDigest string,
) error {
	if repo == nil || repo.db == nil {
		return gorm.ErrInvalidData
	}
	return createWorkerSpecDependencyArtifact(
		repo.db.WithContext(ctx),
		organizationID,
		snapshotID,
		artifactJSON,
		artifactDigest,
	)
}

func (repo *workerSpecDependencyArtifactRepo) Delete(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
) error {
	if repo == nil || repo.db == nil || organizationID <= 0 || snapshotID <= 0 {
		return gorm.ErrInvalidData
	}
	return repo.db.WithContext(ctx).
		Where("organization_id = ? AND worker_spec_snapshot_id = ?", organizationID, snapshotID).
		Delete(&workerSpecDependencyArtifactRecord{}).Error
}

func (repo *workerSpecDependencyArtifactRepo) GetBySnapshotID(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
) (workerdependency.Document, error) {
	if repo == nil || repo.db == nil || organizationID <= 0 || snapshotID <= 0 {
		return workerdependency.Document{}, gorm.ErrInvalidData
	}
	var record workerSpecDependencyArtifactRecord
	err := repo.db.WithContext(ctx).
		Where("organization_id = ? AND worker_spec_snapshot_id = ?", organizationID, snapshotID).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return workerdependency.Document{}, err
	}
	if err != nil {
		return workerdependency.Document{}, err
	}
	document, err := workerdependency.Decode(record.ArtifactJSON)
	if err != nil {
		return workerdependency.Document{}, err
	}
	digest, err := workerdependency.Digest(document)
	if err != nil {
		return workerdependency.Document{}, err
	}
	if document.OrganizationID != organizationID || digest != record.ArtifactDigest {
		return workerdependency.Document{}, fmt.Errorf("worker dependency artifact binding mismatch")
	}
	return document, nil
}
