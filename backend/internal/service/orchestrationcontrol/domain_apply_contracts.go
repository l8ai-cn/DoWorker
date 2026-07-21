package orchestrationcontrol

import (
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

type BindingApplyBuilder = ApplyBuilder

type WorkerTemplateApplyBuilder func(
	LockedApplyState,
	int64,
) (ApplyMutation, error)

type AppliedWorkerTemplate struct {
	Head                 control.ResourceHead
	WorkerSpecSnapshotID int64
}

type ExpertApplyProjection struct {
	Name                 string
	Description          string
	Category             string
	ReleaseNotes         string
	Prompt               string
	WorkerSpecSnapshotID int64
}

type ExpertApplyMutation struct {
	ApplyMutation
	Projection ExpertApplyProjection
}

type ExpertApplyBuilder func(
	LockedApplyState,
) (ExpertApplyMutation, error)

type AppliedExpert struct {
	Head                 control.ResourceHead
	ExpertID             int64
	WorkerSpecSnapshotID int64
	ResourceRevision     int64
}

type WorkflowApplyProjection struct {
	Name                 string
	Prompt               string
	ExecutionMode        string
	CronExpression       string
	SandboxStrategy      string
	SessionPersistence   bool
	ConcurrencyPolicy    string
	MaxConcurrentRuns    int
	MaxRetainedRuns      int
	TimeoutMinutes       int
	IdleTimeoutSeconds   int
	CallbackURL          string
	WorkerSpecSnapshotID int64
	NextRunAt            *time.Time
}

type WorkflowApplyMutation struct {
	ApplyMutation
	Projection WorkflowApplyProjection
}

type WorkflowApplyBuilder func(
	LockedApplyState,
) (WorkflowApplyMutation, error)

type AppliedWorkflow struct {
	Head                 control.ResourceHead
	WorkflowID           int64
	WorkerSpecSnapshotID int64
	ResourceRevision     int64
}

type GoalLoopApplyProjection struct {
	Name                 string
	Description          string
	Objective            string
	AcceptanceCriteria   []string
	VerificationCommand  string
	MaxIterations        int
	TokenBudget          *int64
	TimeoutMinutes       int
	NoProgressLimit      int
	SameErrorLimit       int
	EscalationPolicy     string
	WorkerSpecSnapshotID int64
}

type GoalLoopApplyMutation struct {
	ApplyMutation
	Projection GoalLoopApplyProjection
}

type GoalLoopApplyBuilder func(
	LockedApplyState,
) (GoalLoopApplyMutation, error)

type AppliedGoalLoop struct {
	Head                 control.ResourceHead
	GoalLoopID           int64
	WorkerSpecSnapshotID int64
	ResourceRevision     int64
}
