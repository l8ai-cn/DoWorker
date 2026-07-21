package infra

import (
	"context"
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) GetRevision(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	resourceID, revision int64,
) (orchestrationcontrol.ResourceRevision, error) {
	if err := validateRevisionQuery(scope, resourceID, revision); err != nil {
		return orchestrationcontrol.ResourceRevision{}, err
	}
	var record orchestrationRevisionRecord
	err := repo.db.WithContext(ctx).
		Where(
			"organization_id = ? AND resource_id = ? AND revision = ?",
			scope.OrganizationID,
			resourceID,
			revision,
		).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationcontrol.ResourceRevision{}, orchestrationcontrol.ErrNotFound
	}
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, err
	}
	return repo.revisionDomain(ctx, scope, record)
}

func (repo *orchestrationResourceRepo) ListRevisions(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	resourceID int64,
	limit, offset int,
) ([]orchestrationcontrol.ResourceRevision, error) {
	if err := validateRevisionQuery(scope, resourceID, 1); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, orchestrationcontrol.ErrInvalid
	}
	var records []orchestrationRevisionRecord
	err := repo.db.WithContext(ctx).
		Where(
			"organization_id = ? AND resource_id = ?",
			scope.OrganizationID,
			resourceID,
		).
		Order("revision DESC").Limit(limit).Offset(offset).Find(&records).Error
	if err != nil {
		return nil, err
	}
	revisions := make([]orchestrationcontrol.ResourceRevision, 0, len(records))
	for _, record := range records {
		revision, err := repo.revisionDomain(ctx, scope, record)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, revision)
	}
	return revisions, nil
}

func validateRevisionQuery(
	scope orchestrationcontrol.Scope,
	resourceID, revision int64,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if resourceID <= 0 || revision <= 0 {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}
