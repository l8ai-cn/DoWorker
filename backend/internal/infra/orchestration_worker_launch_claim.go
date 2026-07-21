package infra

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) ClaimWorkerLaunch(
	ctx context.Context,
	scope control.Scope,
	launchID int64,
	leaseDuration time.Duration,
	claimToken string,
) (controlservice.WorkerLaunchClaim, error) {
	if err := validateWorkerLaunchClaimRequest(
		scope,
		launchID,
		leaseDuration,
		claimToken,
	); err != nil {
		return controlservice.WorkerLaunchClaim{}, err
	}
	var claimed orchestrationWorkerLaunchRecord
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now, err := orchestrationTransactionTime(tx)
		if err != nil {
			return err
		}
		result := tx.Raw(`
UPDATE orchestration_worker_launches
SET state = ?, claim_token = ?, lease_expires_at = ?,
	attempt_count = attempt_count + 1, last_error = NULL, updated_at = ?
WHERE organization_id = ? AND id = ?
	AND (
		state = ?
		OR (state = ? AND lease_expires_at <= ?)
	)
RETURNING *`,
			workerLaunchStateMaterializing,
			claimToken,
			now.Add(leaseDuration),
			now,
			scope.OrganizationID,
			launchID,
			workerLaunchStatePending,
			workerLaunchStateMaterializing,
			now,
		).Scan(&claimed)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			return nil
		}
		return workerLaunchClaimConflict(tx, scope, launchID)
	})
	if err != nil {
		return controlservice.WorkerLaunchClaim{}, err
	}
	return workerLaunchClaimFromRecord(claimed)
}

func validateWorkerLaunchClaimRequest(
	scope control.Scope,
	launchID int64,
	leaseDuration time.Duration,
	claimToken string,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	parsed, err := uuid.Parse(claimToken)
	if launchID <= 0 || leaseDuration <= 0 ||
		leaseDuration > 10*time.Minute || err != nil ||
		parsed == uuid.Nil || parsed.String() != claimToken {
		return control.ErrInvalid
	}
	return nil
}

func workerLaunchClaimConflict(
	tx *gorm.DB,
	scope control.Scope,
	launchID int64,
) error {
	var current orchestrationWorkerLaunchRecord
	err := tx.Where(
		"organization_id = ? AND id = ?",
		scope.OrganizationID,
		launchID,
	).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return control.ErrNotFound
	}
	if err != nil {
		return err
	}
	switch current.State {
	case workerLaunchStateMaterializing:
		return controlservice.ErrWorkerLaunchInProgress
	case workerLaunchStateDispatched:
		return control.ErrConsumed
	default:
		return control.ErrCorrupt
	}
}
