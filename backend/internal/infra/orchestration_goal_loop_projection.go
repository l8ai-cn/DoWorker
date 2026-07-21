package infra

import (
	"encoding/json"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

type orchestrationGoalLoopRecord struct {
	ID                            int64           `gorm:"column:id;primaryKey"`
	OrganizationID                int64           `gorm:"column:organization_id"`
	CreatedByID                   int64           `gorm:"column:created_by_id"`
	Name                          string          `gorm:"column:name"`
	Slug                          string          `gorm:"column:slug"`
	Description                   *string         `gorm:"column:description"`
	WorkerSpecSnapshotID          int64           `gorm:"column:worker_spec_snapshot_id"`
	Objective                     string          `gorm:"column:objective"`
	AcceptanceCriteria            json.RawMessage `gorm:"column:acceptance_criteria;type:jsonb"`
	VerificationCommand           string          `gorm:"column:verification_command"`
	Status                        string          `gorm:"column:status"`
	MaxIterations                 int             `gorm:"column:max_iterations"`
	TokenBudget                   *int64          `gorm:"column:token_budget"`
	TimeoutMinutes                int             `gorm:"column:timeout_minutes"`
	NoProgressLimit               int             `gorm:"column:no_progress_limit"`
	SameErrorLimit                int             `gorm:"column:same_error_limit"`
	EscalationPolicy              string          `gorm:"column:escalation_policy"`
	OrchestrationResourceID       int64           `gorm:"column:orchestration_resource_id"`
	OrchestrationResourceRevision int64           `gorm:"column:orchestration_resource_revision"`
	CreatedAt                     time.Time       `gorm:"column:created_at"`
	UpdatedAt                     time.Time       `gorm:"column:updated_at"`
}

func (orchestrationGoalLoopRecord) TableName() string {
	return "goal_loops"
}

func writeGoalLoopProjection(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
	mutation controlservice.GoalLoopApplyMutation,
) (int64, error) {
	criteria, err := control.CanonicalJSONArray(
		mutation.Projection.AcceptanceCriteria,
	)
	if err != nil {
		return 0, control.ErrCorrupt
	}
	projection := mutation.Projection
	record := orchestrationGoalLoopRecord{
		OrganizationID:                state.Plan.Scope.OrganizationID,
		CreatedByID:                   state.Plan.ActorID,
		Name:                          projection.Name,
		Slug:                          state.Plan.Target.Name.String(),
		Description:                   optionalProjectionText(projection.Description),
		WorkerSpecSnapshotID:          projection.WorkerSpecSnapshotID,
		Objective:                     projection.Objective,
		AcceptanceCriteria:            criteria,
		VerificationCommand:           projection.VerificationCommand,
		Status:                        goalloop.StatusDraft,
		MaxIterations:                 projection.MaxIterations,
		TokenBudget:                   projection.TokenBudget,
		TimeoutMinutes:                projection.TimeoutMinutes,
		NoProgressLimit:               projection.NoProgressLimit,
		SameErrorLimit:                projection.SameErrorLimit,
		EscalationPolicy:              projection.EscalationPolicy,
		OrchestrationResourceID:       mutation.Head.ID,
		OrchestrationResourceRevision: mutation.Head.Revision,
		CreatedAt:                     state.AppliedAt, UpdatedAt: state.AppliedAt,
	}
	if err := tx.Create(&record).Error; err != nil {
		return 0, err
	}
	return record.ID, nil
}
