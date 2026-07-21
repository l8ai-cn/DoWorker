package infra

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

const (
	maxWorkerDispatchTTL    = 7 * 24 * time.Hour
	maxWorkerCommandPayload = 4 << 20
)

type workerLaunchPodRecord struct {
	ID                   int64  `gorm:"column:id"`
	OrganizationID       int64  `gorm:"column:organization_id"`
	PodKey               string `gorm:"column:pod_key"`
	RunnerID             int64  `gorm:"column:runner_id"`
	RunnerOrganizationID int64  `gorm:"column:runner_organization_id"`
	CreatedByID          int64  `gorm:"column:created_by_id"`
	WorkerSpecSnapshotID *int64 `gorm:"column:worker_spec_snapshot_id"`
	WorkerLaunchID       *int64 `gorm:"column:orchestration_worker_launch_id"`
	Status               string `gorm:"column:status"`
}

func validateWorkerLaunchCompletion(
	scope control.Scope,
	claim controlservice.WorkerLaunchClaim,
	launch controlservice.WorkerPodLaunch,
	dispatchTTL time.Duration,
) error {
	if err := validateWorkerLaunchClaimCoordinates(scope, claim); err != nil {
		return err
	}
	token, err := uuid.Parse(claim.ClaimToken)
	if err != nil || token == uuid.Nil ||
		token.String() != claim.ClaimToken ||
		launch.PodID <= 0 || launch.PodKey == "" ||
		launch.RunnerID <= 0 || len(launch.CommandPayload) == 0 ||
		len(launch.CommandPayload) > maxWorkerCommandPayload ||
		dispatchTTL <= 0 || dispatchTTL > maxWorkerDispatchTTL {
		return control.ErrInvalid
	}
	return nil
}

func lockWorkerLaunch(
	tx *gorm.DB,
	scope control.Scope,
	launchID int64,
) (orchestrationWorkerLaunchRecord, error) {
	var current orchestrationWorkerLaunchRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		"organization_id = ? AND id = ?",
		scope.OrganizationID,
		launchID,
	).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return current, control.ErrNotFound
	}
	return current, err
}

func verifyWorkerLaunchLease(
	current orchestrationWorkerLaunchRecord,
	claim controlservice.WorkerLaunchClaim,
	now time.Time,
) error {
	if current.State != workerLaunchStateMaterializing ||
		current.ClaimToken == nil ||
		*current.ClaimToken != claim.ClaimToken ||
		current.LeaseExpiresAt == nil ||
		!now.Before(*current.LeaseExpiresAt) {
		return controlservice.ErrWorkerLaunchLeaseLost
	}
	if current.OrganizationID != claim.OrganizationID ||
		current.ActorID != claim.ActorID ||
		current.PlanID != claim.PlanID ||
		current.ResourceID != claim.ResourceID ||
		current.ResourceRevision != claim.ResourceRevision ||
		current.WorkerSpecSnapshotID != claim.WorkerSpecSnapshotID ||
		!current.LeaseExpiresAt.Equal(claim.LeaseExpiresAt) {
		return control.ErrCorrupt
	}
	return nil
}

func verifyWorkerLaunchPod(
	tx *gorm.DB,
	scope control.Scope,
	claim controlservice.WorkerLaunchClaim,
	launch controlservice.WorkerPodLaunch,
) error {
	var pod workerLaunchPodRecord
	err := tx.Table("pods AS p").
		Select(`
p.id, p.organization_id, p.pod_key, p.runner_id,
r.organization_id AS runner_organization_id, p.created_by_id,
p.worker_spec_snapshot_id, p.orchestration_worker_launch_id, p.status`).
		Joins("JOIN runners AS r ON r.id = p.runner_id").
		Where(
			"p.organization_id = ? AND p.id = ? AND p.pod_key = ?",
			scope.OrganizationID,
			launch.PodID,
			launch.PodKey,
		).
		Take(&pod).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return control.ErrCorrupt
	}
	if err != nil {
		return err
	}
	if pod.OrganizationID != scope.OrganizationID ||
		pod.RunnerOrganizationID != scope.OrganizationID ||
		pod.RunnerID != launch.RunnerID ||
		pod.CreatedByID != claim.ActorID ||
		pod.WorkerSpecSnapshotID == nil ||
		*pod.WorkerSpecSnapshotID != claim.WorkerSpecSnapshotID ||
		pod.WorkerLaunchID == nil ||
		*pod.WorkerLaunchID != claim.LaunchID ||
		pod.Status != agentpod.StatusQueued {
		return control.ErrCorrupt
	}
	return nil
}
