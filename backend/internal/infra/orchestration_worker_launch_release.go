package infra

import (
	"context"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) ReleaseWorkerLaunch(
	ctx context.Context,
	scope control.Scope,
	claim controlservice.WorkerLaunchClaim,
	reason string,
) error {
	if err := validateWorkerLaunchClaimCoordinates(scope, claim); err != nil {
		return err
	}
	if reason == "" || len(reason) > 500 {
		return control.ErrInvalid
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now, err := orchestrationTransactionTime(tx)
		if err != nil {
			return err
		}
		result := tx.Model(&orchestrationWorkerLaunchRecord{}).Where(
			"organization_id = ? AND id = ? AND state = ? AND claim_token = ?",
			scope.OrganizationID,
			claim.LaunchID,
			workerLaunchStateMaterializing,
			claim.ClaimToken,
		).Updates(map[string]any{
			"state":            workerLaunchStatePending,
			"claim_token":      nil,
			"lease_expires_at": nil,
			"last_error":       reason,
			"updated_at":       now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return controlservice.ErrWorkerLaunchLeaseLost
		}
		return nil
	})
}

func validateWorkerLaunchClaimCoordinates(
	scope control.Scope,
	claim controlservice.WorkerLaunchClaim,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if claim.LaunchID <= 0 ||
		claim.OrganizationID != scope.OrganizationID ||
		claim.ActorID != scope.ActorID ||
		claim.ResourceID <= 0 || claim.ResourceRevision <= 0 ||
		claim.WorkerSpecSnapshotID <= 0 ||
		claim.ClaimToken == "" || claim.LeaseExpiresAt.IsZero() {
		return control.ErrInvalid
	}
	return nil
}
