package orchestrationcontrol

import (
	"errors"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

var (
	ErrWorkerLaunchInProgress = errors.New(
		"worker launch is already being materialized",
	)
	ErrWorkerLaunchLeaseLost = errors.New(
		"worker launch lease is no longer owned by this request",
	)
)

type WorkerLaunchProjection struct {
	WorkerSpecSnapshotID int64
	Prompt               *string
	Alias                string
}

type WorkerApplyMutation struct {
	ApplyMutation
	Launch WorkerLaunchProjection
}

type WorkerApplyBuilder func(
	LockedApplyState,
) (WorkerApplyMutation, error)

type AppliedWorker struct {
	Head                 control.ResourceHead
	LaunchID             int64
	WorkerSpecSnapshotID int64
	ResourceRevision     int64
	PodID                int64
	PodKey               string
	RunnerID             int64
}

type WorkerLaunchClaim struct {
	LaunchID             int64
	PlanID               string
	OrganizationID       int64
	ActorID              int64
	ResourceID           int64
	ResourceRevision     int64
	WorkerSpecSnapshotID int64
	Prompt               *string
	Alias                string
	ClaimToken           string
	LeaseExpiresAt       time.Time
}

type WorkerPodLaunch struct {
	PodID          int64
	PodKey         string
	RunnerID       int64
	CommandPayload []byte
}
