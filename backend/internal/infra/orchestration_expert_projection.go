package infra

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type orchestrationExpertRecord struct {
	ID                            int64           `gorm:"column:id;primaryKey"`
	OrganizationID                int64           `gorm:"column:organization_id"`
	Slug                          string          `gorm:"column:slug"`
	Name                          string          `gorm:"column:name"`
	Description                   *string         `gorm:"column:description"`
	AgentSlug                     string          `gorm:"column:agent_slug"`
	Prompt                        *string         `gorm:"column:prompt"`
	InteractionMode               string          `gorm:"column:interaction_mode"`
	AutomationLevel               string          `gorm:"column:automation_level"`
	SkillSlugs                    pq.StringArray  `gorm:"column:skill_slugs;type:text[]"`
	WorkerSpecSnapshotID          int64           `gorm:"column:worker_spec_snapshot_id"`
	Metadata                      json.RawMessage `gorm:"column:metadata;type:jsonb"`
	CreatedByID                   int64           `gorm:"column:created_by_id"`
	OrchestrationResourceID       int64           `gorm:"column:orchestration_resource_id"`
	OrchestrationResourceRevision int64           `gorm:"column:orchestration_resource_revision"`
	CreatedAt                     time.Time       `gorm:"column:created_at"`
	UpdatedAt                     time.Time       `gorm:"column:updated_at"`
}

func (orchestrationExpertRecord) TableName() string {
	return "experts"
}

func writeExpertProjection(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
	mutation controlservice.ExpertApplyMutation,
) (int64, error) {
	record, err := expertProjectionRecord(tx, state, mutation)
	if err != nil {
		return 0, err
	}
	if state.Head == nil {
		if err := tx.Create(&record).Error; err != nil {
			return 0, err
		}
		return record.ID, nil
	}
	var current orchestrationExpertRecord
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
	result := tx.Model(&orchestrationExpertRecord{}).Where(
		"id = ? AND organization_id = ? AND orchestration_resource_revision = ?",
		current.ID,
		state.Plan.Scope.OrganizationID,
		state.Head.Revision,
	).Updates(map[string]any{
		"name": record.Name, "description": record.Description,
		"prompt": record.Prompt, "metadata": record.Metadata,
		"agent_slug": record.AgentSlug, "interaction_mode": record.InteractionMode,
		"automation_level": record.AutomationLevel, "skill_slugs": record.SkillSlugs,
		"worker_spec_snapshot_id":         record.WorkerSpecSnapshotID,
		"orchestration_resource_revision": record.OrchestrationResourceRevision,
		"updated_at":                      record.UpdatedAt,
	})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected != 1 {
		return 0, control.ErrStale
	}
	return current.ID, nil
}

func expertProjectionRecord(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
	mutation controlservice.ExpertApplyMutation,
) (orchestrationExpertRecord, error) {
	runtime, err := loadExpertProjectionRuntime(
		tx,
		state.Plan.Scope.OrganizationID,
		mutation.Projection.WorkerSpecSnapshotID,
	)
	if err != nil {
		return orchestrationExpertRecord{}, err
	}
	metadata, err := control.CanonicalJSONObject(map[string]string{
		"category":     mutation.Projection.Category,
		"releaseNotes": mutation.Projection.ReleaseNotes,
	})
	if err != nil {
		return orchestrationExpertRecord{}, control.ErrCorrupt
	}
	return orchestrationExpertRecord{
		OrganizationID:                state.Plan.Scope.OrganizationID,
		Slug:                          state.Plan.Target.Name.String(),
		Name:                          mutation.Projection.Name,
		Description:                   optionalProjectionText(mutation.Projection.Description),
		AgentSlug:                     runtime.AgentSlug,
		Prompt:                        optionalProjectionText(mutation.Projection.Prompt),
		InteractionMode:               runtime.InteractionMode,
		AutomationLevel:               runtime.AutomationLevel,
		SkillSlugs:                    runtime.SkillSlugs,
		WorkerSpecSnapshotID:          mutation.Projection.WorkerSpecSnapshotID,
		Metadata:                      metadata,
		CreatedByID:                   state.Plan.ActorID,
		OrchestrationResourceID:       mutation.Head.ID,
		OrchestrationResourceRevision: mutation.Head.Revision,
		CreatedAt:                     state.AppliedAt, UpdatedAt: state.AppliedAt,
	}, nil
}

func optionalProjectionText(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
