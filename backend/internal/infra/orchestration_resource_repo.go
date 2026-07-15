package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

var _ orchestrationservice.Repository = (*orchestrationResourceRepo)(nil)

type orchestrationResourceRepo struct {
	db *gorm.DB
}

func NewOrchestrationResourceRepository(
	db *gorm.DB,
) *orchestrationResourceRepo {
	return &orchestrationResourceRepo{db: db}
}

func (repo *orchestrationResourceRepo) GetResource(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	target orchestrationcontrol.ResourceTarget,
) (orchestrationcontrol.ResourceHead, error) {
	if err := target.Validate(scope); err != nil {
		return orchestrationcontrol.ResourceHead{}, err
	}
	var record orchestrationResourceRecord
	err := repo.db.WithContext(ctx).
		Where(resourceIdentityQuery(scope, target)).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationcontrol.ResourceHead{}, orchestrationcontrol.ErrNotFound
	}
	if err != nil {
		return orchestrationcontrol.ResourceHead{}, err
	}
	return record.domain(scope)
}

func (repo *orchestrationResourceRepo) ListResources(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	filter orchestrationservice.ResourceListFilter,
) (orchestrationservice.ResourceListPage, error) {
	if err := validateResourceList(scope, filter); err != nil {
		return orchestrationservice.ResourceListPage{}, err
	}
	query := repo.db.WithContext(ctx).
		Model(&orchestrationResourceRecord{}).
		Where("organization_id = ?", scope.OrganizationID)
	if filter.Kind != "" {
		query = query.Where("kind = ?", filter.Kind)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return orchestrationservice.ResourceListPage{}, err
	}
	var records []orchestrationResourceRecord
	if err := query.Order("kind, namespace, name, id").
		Limit(filter.Limit).Offset(filter.Offset).Find(&records).Error; err != nil {
		return orchestrationservice.ResourceListPage{}, err
	}
	resources := make([]orchestrationcontrol.ResourceHead, 0, len(records))
	for _, record := range records {
		resource, err := record.domain(scope)
		if err != nil {
			return orchestrationservice.ResourceListPage{}, err
		}
		resources = append(resources, resource)
	}
	return orchestrationservice.ResourceListPage{
		Items: resources,
		Total: total,
	}, nil
}

func validateResourceList(
	scope orchestrationcontrol.Scope,
	filter orchestrationservice.ResourceListFilter,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if filter.Limit <= 0 || filter.Limit > 100 || filter.Offset < 0 {
		return orchestrationcontrol.ErrInvalid
	}
	if filter.Kind == "" {
		return nil
	}
	return (orchestrationresource.TypeMeta{
		APIVersion: orchestrationresource.APIVersionV1Alpha1,
		Kind:       filter.Kind,
	}).Validate()
}

func resourceIdentityQuery(
	scope orchestrationcontrol.Scope,
	target orchestrationcontrol.ResourceTarget,
) map[string]any {
	return map[string]any{
		"organization_id": scope.OrganizationID,
		"api_version":     target.APIVersion,
		"kind":            target.Kind,
		"namespace":       target.Namespace.String(),
		"name":            target.Name.String(),
	}
}
