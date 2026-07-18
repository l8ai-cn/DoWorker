package infra

import (
	"context"
	"database/sql"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
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
	if filter.EnvironmentBundle == nil {
		return listOrchestrationResources(
			repo.db.WithContext(ctx),
			scope,
			filter,
		)
	}
	return repo.listEnvironmentBundleResources(ctx, scope, filter)
}

func (repo *orchestrationResourceRepo) listEnvironmentBundleResources(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	filter orchestrationservice.ResourceListFilter,
) (orchestrationservice.ResourceListPage, error) {
	var page orchestrationservice.ResourceListPage
	run := func(tx *gorm.DB) error {
		var err error
		page, err = listOrchestrationResources(tx, scope, filter)
		return err
	}
	switch repo.db.Dialector.Name() {
	case "postgres":
		err := repo.db.WithContext(ctx).Transaction(run, &sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
			ReadOnly:  true,
		})
		return page, err
	case "sqlite":
		err := repo.db.WithContext(ctx).Transaction(run)
		return page, err
	default:
		return orchestrationservice.ResourceListPage{}, orchestrationservice.ErrUnavailable
	}
}

func listOrchestrationResources(
	db *gorm.DB,
	scope orchestrationcontrol.Scope,
	filter orchestrationservice.ResourceListFilter,
) (orchestrationservice.ResourceListPage, error) {
	query := db.
		Model(&orchestrationResourceRecord{}).
		Where("orchestration_resources.organization_id = ?", scope.OrganizationID)
	if filter.Kind != "" {
		query = query.Where("orchestration_resources.kind = ?", filter.Kind)
	}
	query, err := filterEnvironmentBundleReferences(query, scope, filter)
	if err != nil {
		return orchestrationservice.ResourceListPage{}, err
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return orchestrationservice.ResourceListPage{}, err
	}
	var records []orchestrationResourceRecord
	if err := query.Order(
		"orchestration_resources.kind, orchestration_resources.namespace, " +
			"orchestration_resources.name, orchestration_resources.id",
	).
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
	return filter.Validate(scope)
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
