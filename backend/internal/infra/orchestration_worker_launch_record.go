package infra

import (
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

const (
	workerLaunchStatePending       = "pending"
	workerLaunchStateMaterializing = "materializing"
	workerLaunchStateDispatched    = "dispatched"
)

type orchestrationWorkerLaunchRecord struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	OrganizationID       int64      `gorm:"column:organization_id"`
	ActorID              int64      `gorm:"column:actor_id"`
	PlanID               string     `gorm:"column:plan_id"`
	ResourceID           int64      `gorm:"column:resource_id"`
	ResourceRevision     int64      `gorm:"column:resource_revision"`
	WorkerSpecSnapshotID int64      `gorm:"column:worker_spec_snapshot_id"`
	Prompt               *string    `gorm:"column:prompt"`
	Alias                string     `gorm:"column:alias"`
	State                string     `gorm:"column:state"`
	ClaimToken           *string    `gorm:"column:claim_token"`
	LeaseExpiresAt       *time.Time `gorm:"column:lease_expires_at"`
	AttemptCount         int        `gorm:"column:attempt_count"`
	LastError            *string    `gorm:"column:last_error"`
	PodID                *int64     `gorm:"column:pod_id"`
	PodKey               *string    `gorm:"column:pod_key"`
	RunnerID             *int64     `gorm:"column:runner_id"`
	DispatchedAt         *time.Time `gorm:"column:dispatched_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (orchestrationWorkerLaunchRecord) TableName() string {
	return "orchestration_worker_launches"
}

func newWorkerLaunchRecord(
	state control.Scope,
	planID string,
	appliedAt time.Time,
	mutation controlservice.WorkerApplyMutation,
) orchestrationWorkerLaunchRecord {
	return orchestrationWorkerLaunchRecord{
		OrganizationID:       state.OrganizationID,
		ActorID:              state.ActorID,
		PlanID:               planID,
		ResourceID:           mutation.Head.ID,
		ResourceRevision:     mutation.Head.Revision,
		WorkerSpecSnapshotID: mutation.Launch.WorkerSpecSnapshotID,
		Prompt:               cloneOptionalString(mutation.Launch.Prompt),
		Alias:                mutation.Launch.Alias,
		State:                workerLaunchStatePending,
		CreatedAt:            appliedAt,
		UpdatedAt:            appliedAt,
	}
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func workerLaunchClaimFromRecord(
	record orchestrationWorkerLaunchRecord,
) (controlservice.WorkerLaunchClaim, error) {
	if record.ID <= 0 || record.OrganizationID <= 0 ||
		record.ActorID <= 0 || record.PlanID == "" ||
		record.ResourceID <= 0 || record.ResourceRevision <= 0 ||
		record.WorkerSpecSnapshotID <= 0 ||
		record.State != workerLaunchStateMaterializing ||
		record.ClaimToken == nil || *record.ClaimToken == "" ||
		record.LeaseExpiresAt == nil ||
		record.LeaseExpiresAt.IsZero() {
		return controlservice.WorkerLaunchClaim{}, control.ErrCorrupt
	}
	return controlservice.WorkerLaunchClaim{
		LaunchID: record.ID, PlanID: record.PlanID,
		OrganizationID: record.OrganizationID,
		ActorID:        record.ActorID, ResourceID: record.ResourceID,
		ResourceRevision:     record.ResourceRevision,
		WorkerSpecSnapshotID: record.WorkerSpecSnapshotID,
		Prompt:               cloneOptionalString(record.Prompt), Alias: record.Alias,
		ClaimToken:     *record.ClaimToken,
		LeaseExpiresAt: record.LeaseExpiresAt.UTC(),
	}, nil
}
