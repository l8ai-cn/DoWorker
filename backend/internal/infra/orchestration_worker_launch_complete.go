package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) CompleteWorkerLaunch(
	ctx context.Context,
	scope control.Scope,
	claim controlservice.WorkerLaunchClaim,
	launch controlservice.WorkerPodLaunch,
	dispatchTTL time.Duration,
) (controlservice.AppliedWorker, error) {
	if err := validateWorkerLaunchCompletion(
		scope,
		claim,
		launch,
		dispatchTTL,
	); err != nil {
		return controlservice.AppliedWorker{}, err
	}
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now, err := orchestrationTransactionTime(tx)
		if err != nil {
			return err
		}
		current, err := lockWorkerLaunch(tx, scope, claim.LaunchID)
		if err != nil {
			return err
		}
		if err := verifyWorkerLaunchLease(current, claim, now); err != nil {
			return err
		}
		if err := verifyWorkerLaunchPod(tx, scope, claim, launch); err != nil {
			return err
		}
		command := agentpod.PendingCommand{
			OrganizationID: scope.OrganizationID,
			RunnerID:       launch.RunnerID,
			PodKey:         launch.PodKey,
			CommandType:    agentpod.CommandTypeCreatePod,
			CommandID:      fmt.Sprintf("worker-%d", claim.LaunchID),
			Payload:        append([]byte(nil), launch.CommandPayload...),
			ExpiresAt:      now.Add(dispatchTTL),
			CreatedAt:      now,
		}
		if err := tx.Create(&command).Error; err != nil {
			if isUniqueViolation(err) {
				return control.ErrCorrupt
			}
			return err
		}
		result := tx.Model(&orchestrationWorkerLaunchRecord{}).Where(
			"organization_id = ? AND id = ? AND state = ? AND claim_token = ?",
			scope.OrganizationID,
			claim.LaunchID,
			workerLaunchStateMaterializing,
			claim.ClaimToken,
		).Updates(map[string]any{
			"state":            workerLaunchStateDispatched,
			"claim_token":      nil,
			"lease_expires_at": nil,
			"pod_id":           launch.PodID,
			"pod_key":          launch.PodKey,
			"runner_id":        launch.RunnerID,
			"dispatched_at":    now,
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
	if err != nil {
		return controlservice.AppliedWorker{}, err
	}
	return repo.loadAppliedWorker(ctx, scope, claim.PlanID)
}
