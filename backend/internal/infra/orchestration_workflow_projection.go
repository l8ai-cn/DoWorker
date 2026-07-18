package infra

import (
	"encoding/json"
	"errors"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type orchestrationWorkflowRecord struct {
	ID                            int64           `gorm:"column:id;primaryKey"`
	OrganizationID                int64           `gorm:"column:organization_id"`
	Name                          string          `gorm:"column:name"`
	Slug                          string          `gorm:"column:slug"`
	AgentSlug                     string          `gorm:"column:agent_slug"`
	PermissionMode                string          `gorm:"column:permission_mode"`
	PromptTemplate                string          `gorm:"column:prompt_template"`
	UsedEnvBundles                pq.StringArray  `gorm:"column:used_env_bundles;type:text[]"`
	ConfigOverrides               json.RawMessage `gorm:"column:config_overrides;type:jsonb"`
	PromptVariables               json.RawMessage `gorm:"column:prompt_variables;type:jsonb"`
	ExecutionMode                 string          `gorm:"column:execution_mode"`
	CronExpression                *string         `gorm:"column:cron_expression"`
	AutopilotConfig               json.RawMessage `gorm:"column:autopilot_config;type:jsonb"`
	CallbackURL                   *string         `gorm:"column:callback_url"`
	Status                        string          `gorm:"column:status"`
	SandboxStrategy               string          `gorm:"column:sandbox_strategy"`
	SessionPersistence            bool            `gorm:"column:session_persistence"`
	ConcurrencyPolicy             string          `gorm:"column:concurrency_policy"`
	MaxConcurrentRuns             int             `gorm:"column:max_concurrent_runs"`
	MaxRetainedRuns               int             `gorm:"column:max_retained_runs"`
	TimeoutMinutes                int             `gorm:"column:timeout_minutes"`
	IdleTimeoutSec                int             `gorm:"column:idle_timeout_sec"`
	CreatedByID                   int64           `gorm:"column:created_by_id"`
	NextRunAt                     *time.Time      `gorm:"column:next_run_at"`
	WorkerSpecSnapshotID          int64           `gorm:"column:worker_spec_snapshot_id"`
	OrchestrationResourceID       int64           `gorm:"column:orchestration_resource_id"`
	OrchestrationResourceRevision int64           `gorm:"column:orchestration_resource_revision"`
	CreatedAt                     time.Time       `gorm:"column:created_at"`
	UpdatedAt                     time.Time       `gorm:"column:updated_at"`
}

func (orchestrationWorkflowRecord) TableName() string {
	return "workflows"
}

func writeWorkflowProjection(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
	mutation controlservice.WorkflowApplyMutation,
) (int64, error) {
	record, err := workflowProjectionRecord(state, mutation)
	if err != nil {
		return 0, err
	}
	if state.Head == nil {
		if err := tx.Create(&record).Error; err != nil {
			return 0, err
		}
		return record.ID, nil
	}
	var current orchestrationWorkflowRecord
	err = tx.Where(
		"organization_id = ? AND orchestration_resource_id = ?",
		state.Plan.Scope.OrganizationID,
		state.Head.ID,
	).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, control.ErrCorrupt
	}
	if err != nil {
		return 0, err
	}
	if current.OrchestrationResourceRevision != state.Head.Revision {
		return 0, control.ErrCorrupt
	}
	result := tx.Model(&orchestrationWorkflowRecord{}).Where(
		"id = ? AND organization_id = ? AND orchestration_resource_revision = ?",
		current.ID,
		state.Plan.Scope.OrganizationID,
		state.Head.Revision,
	).Updates(workflowProjectionUpdates(record))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected != 1 {
		return 0, control.ErrStale
	}
	return current.ID, nil
}

func workflowProjectionRecord(
	state controlservice.LockedApplyState,
	mutation controlservice.WorkflowApplyMutation,
) (orchestrationWorkflowRecord, error) {
	emptyObject, err := control.CanonicalJSONObject(map[string]string{})
	if err != nil {
		return orchestrationWorkflowRecord{}, control.ErrCorrupt
	}
	projection := mutation.Projection
	return orchestrationWorkflowRecord{
		OrganizationID: state.Plan.Scope.OrganizationID,
		Name:           projection.Name, Slug: state.Plan.Target.Name.String(),
		AgentSlug: "resource-native", PermissionMode: "bypassPermissions",
		PromptTemplate: projection.Prompt, UsedEnvBundles: pq.StringArray{},
		ConfigOverrides: emptyObject, PromptVariables: emptyObject,
		ExecutionMode:   projection.ExecutionMode,
		CronExpression:  optionalProjectionText(projection.CronExpression),
		AutopilotConfig: emptyObject,
		CallbackURL:     optionalProjectionText(projection.CallbackURL),
		Status:          projection.Status, SandboxStrategy: projection.SandboxStrategy,
		SessionPersistence: projection.SessionPersistence,
		ConcurrencyPolicy:  projection.ConcurrencyPolicy,
		MaxConcurrentRuns:  projection.MaxConcurrentRuns,
		MaxRetainedRuns:    projection.MaxRetainedRuns,
		TimeoutMinutes:     projection.TimeoutMinutes,
		IdleTimeoutSec:     projection.IdleTimeoutSeconds,
		CreatedByID:        state.Plan.ActorID, NextRunAt: projection.NextRunAt,
		WorkerSpecSnapshotID:          projection.WorkerSpecSnapshotID,
		OrchestrationResourceID:       mutation.Head.ID,
		OrchestrationResourceRevision: mutation.Head.Revision,
		CreatedAt:                     state.AppliedAt, UpdatedAt: state.AppliedAt,
	}, nil
}

func workflowProjectionUpdates(record orchestrationWorkflowRecord) map[string]any {
	return map[string]any{
		"name": record.Name, "prompt_template": record.PromptTemplate,
		"status":              record.Status,
		"execution_mode":      record.ExecutionMode,
		"cron_expression":     record.CronExpression,
		"sandbox_strategy":    record.SandboxStrategy,
		"session_persistence": record.SessionPersistence,
		"concurrency_policy":  record.ConcurrencyPolicy,
		"max_concurrent_runs": record.MaxConcurrentRuns,
		"max_retained_runs":   record.MaxRetainedRuns,
		"timeout_minutes":     record.TimeoutMinutes,
		"idle_timeout_sec":    record.IdleTimeoutSec,
		"callback_url":        record.CallbackURL, "next_run_at": record.NextRunAt,
		"worker_spec_snapshot_id":         record.WorkerSpecSnapshotID,
		"orchestration_resource_revision": record.OrchestrationResourceRevision,
		"last_pod_key":                    nil,
		"sandbox_path":                    nil,
		"updated_at":                      record.UpdatedAt,
	}
}
