package infra

import (
	"errors"
	"strconv"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func loadLockedApplyState(
	tx *gorm.DB,
	scope orchestrationcontrol.Scope,
	planID string,
	appliedAt time.Time,
) (orchestrationservice.LockedApplyState, error) {
	var planRecord orchestrationPlanRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		"id = ? AND organization_id = ? AND actor_id = ?",
		planID,
		scope.OrganizationID,
		scope.ActorID,
	).First(&planRecord).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrNotFound
	}
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	plan, err := planRecord.domain(scope)
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	if plan.Status != orchestrationcontrol.PlanStatusPending {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrConsumed
	}
	if appliedAt.Before(plan.CreatedAt) {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrInvalid
	}
	if !appliedAt.Before(plan.ExpiresAt) {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrExpired
	}
	if err := lockApplyTarget(tx, plan); err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	return loadApplyTarget(tx, scope, plan, appliedAt)
}

func loadApplyTarget(
	tx *gorm.DB,
	scope orchestrationcontrol.Scope,
	plan orchestrationcontrol.Plan,
	appliedAt time.Time,
) (orchestrationservice.LockedApplyState, error) {
	var headRecord orchestrationResourceRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(resourceIdentityQuery(scope, plan.Target)).
		First(&headRecord).Error
	if plan.Operation == orchestrationcontrol.PlanOperationCreate {
		if err == nil {
			return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return orchestrationservice.LockedApplyState{}, err
		}
		return newCreateApplyState(tx, plan, appliedAt)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrStale
	}
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	head, err := headRecord.domain(scope)
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	if head.ID != plan.TargetResourceID || head.Identity.UID != plan.BaseUID ||
		head.ResourceVersion != plan.BaseResourceVersion {
		return orchestrationservice.LockedApplyState{}, orchestrationcontrol.ErrStale
	}
	current, err := loadCurrentRevision(tx, scope, headRecord)
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	return orchestrationservice.LockedApplyState{
		Plan: plan, Head: &head, CurrentRevision: &current,
		ResultResourceID: head.ID, ResultIdentity: head.Identity,
		AppliedAt: appliedAt,
	}, nil
}

func newCreateApplyState(
	tx *gorm.DB,
	plan orchestrationcontrol.Plan,
	appliedAt time.Time,
) (orchestrationservice.LockedApplyState, error) {
	var resourceID int64
	err := tx.Raw(
		"SELECT nextval(pg_get_serial_sequence('orchestration_resources', 'id'))",
	).Scan(&resourceID).Error
	if err != nil {
		return orchestrationservice.LockedApplyState{}, err
	}
	return orchestrationservice.LockedApplyState{
		Plan: plan, ResultResourceID: resourceID,
		AppliedAt: appliedAt,
		ResultIdentity: orchestrationcontrol.ResourceIdentity{
			ResourceTarget: plan.Target,
			UID:            uuid.NewString(),
		},
	}, nil
}

func fmtInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
