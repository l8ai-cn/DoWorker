package infra

import (
	"context"
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) CreatePlan(
	ctx context.Context,
	plan orchestrationcontrol.Plan,
) error {
	record, err := orchestrationPlanRecordFromDomain(plan)
	if err != nil {
		return err
	}
	if plan.Status != orchestrationcontrol.PlanStatusPending {
		return orchestrationcontrol.ErrInvalid
	}
	if err := repo.db.WithContext(ctx).Create(&record).Error; err != nil {
		if isUniqueViolation(err) {
			return orchestrationcontrol.ErrConflict
		}
		return err
	}
	return nil
}

func (repo *orchestrationResourceRepo) GetPlan(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	id string,
) (orchestrationcontrol.Plan, error) {
	if err := scope.Validate(); err != nil {
		return orchestrationcontrol.Plan{}, err
	}
	parsed, err := uuid.Parse(id)
	if err != nil || parsed == uuid.Nil || parsed.String() != id {
		return orchestrationcontrol.Plan{}, orchestrationcontrol.ErrInvalid
	}
	var record orchestrationPlanRecord
	err = repo.db.WithContext(ctx).Where(
		"id = ? AND organization_id = ? AND actor_id = ?",
		id,
		scope.OrganizationID,
		scope.ActorID,
	).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationcontrol.Plan{}, orchestrationcontrol.ErrNotFound
	}
	if err != nil {
		return orchestrationcontrol.Plan{}, err
	}
	return record.domain(scope)
}
