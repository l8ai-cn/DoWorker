package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunApplyTransaction(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	planID string,
	build orchestrationservice.ApplyBuilder,
) (orchestrationcontrol.ResourceHead, error) {
	if err := validateApplyRequest(scope, planID, build); err != nil {
		return orchestrationcontrol.ResourceHead{}, err
	}
	return repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			_ *gorm.DB,
			state orchestrationservice.LockedApplyState,
		) (orchestrationservice.ApplyMutation, error) {
			return build(state)
		},
	)
}

type transactionalApplyBuilder func(
	*gorm.DB,
	orchestrationservice.LockedApplyState,
) (orchestrationservice.ApplyMutation, error)

func (repo *orchestrationResourceRepo) runApplyTransaction(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	planID string,
	build transactionalApplyBuilder,
) (orchestrationcontrol.ResourceHead, error) {
	var applied orchestrationcontrol.ResourceHead
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		appliedAt, err := orchestrationTransactionTime(tx)
		if err != nil {
			return err
		}
		state, err := loadLockedApplyState(tx, scope, planID, appliedAt)
		if errors.Is(err, orchestrationcontrol.ErrConsumed) {
			if err := authorizeConsumedApply(tx, scope, planID); err != nil {
				return err
			}
			return orchestrationcontrol.ErrConsumed
		}
		if err != nil {
			return err
		}
		if err := authorizeLockedApply(tx, scope, state); err != nil {
			return err
		}
		mutation, err := build(tx, state)
		if err != nil {
			return err
		}
		if err := validateApplyMutation(state, mutation); err != nil {
			return err
		}
		if err := writeApplyMutation(tx, state, mutation); err != nil {
			return err
		}
		record, err := orchestrationResourceRecordFromDomain(
			mutation.Head,
			state.Plan.Scope,
		)
		if err != nil {
			return err
		}
		applied, err = record.domain(state.Plan.Scope)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return orchestrationcontrol.ResourceHead{}, mapOrchestrationApplyError(err)
	}
	return applied, nil
}

func validateApplyRequest(
	scope orchestrationcontrol.Scope,
	planID string,
	build orchestrationservice.ApplyBuilder,
) error {
	if build == nil {
		return orchestrationcontrol.ErrInvalid
	}
	return validateApplyCoordinates(scope, planID)
}

func validateApplyCoordinates(
	scope orchestrationcontrol.Scope,
	planID string,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	parsed, err := uuid.Parse(planID)
	if err != nil || parsed == uuid.Nil || parsed.String() != planID {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}

func orchestrationTransactionTime(tx *gorm.DB) (time.Time, error) {
	var appliedAt time.Time
	if err := tx.Raw("SELECT transaction_timestamp()").Scan(&appliedAt).Error; err != nil {
		return time.Time{}, err
	}
	if appliedAt.IsZero() {
		return time.Time{}, orchestrationcontrol.ErrCorrupt
	}
	return appliedAt.UTC(), nil
}

func mapOrchestrationApplyError(err error) error {
	for _, typed := range []error{
		orchestrationcontrol.ErrInvalid,
		orchestrationcontrol.ErrNotFound,
		orchestrationcontrol.ErrConflict,
		orchestrationcontrol.ErrStale,
		orchestrationcontrol.ErrExpired,
		orchestrationcontrol.ErrConsumed,
		orchestrationcontrol.ErrCorrupt,
	} {
		if errors.Is(err, typed) {
			return err
		}
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "can only be consumed once"):
		return orchestrationcontrol.ErrConsumed
	case strings.Contains(message, "plan is stale"):
		return orchestrationcontrol.ErrStale
	case strings.Contains(message, "target already exists"), isUniqueViolation(err):
		return orchestrationcontrol.ErrConflict
	default:
		return err
	}
}

func lockApplyTarget(
	tx *gorm.DB,
	plan orchestrationcontrol.Plan,
) error {
	key := strings.Join([]string{
		fmtInt(plan.Scope.OrganizationID),
		plan.Target.APIVersion,
		plan.Target.Kind,
		plan.Target.Namespace.String(),
		plan.Target.Name.String(),
	}, "|")
	return tx.Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		key,
	).Error
}
