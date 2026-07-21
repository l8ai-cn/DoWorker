package infra

import (
	"encoding/json"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type expertProjectionRuntime struct {
	AgentSlug       string
	InteractionMode string
	AutomationLevel string
	SkillSlugs      pq.StringArray
}

type expertSnapshotSpec struct {
	Runtime struct {
		WorkerType struct {
			Slug string `json:"slug"`
		} `json:"worker_type"`
	} `json:"runtime"`
	TypeConfig struct {
		InteractionMode string `json:"interaction_mode"`
		AutomationLevel string `json:"automation_level"`
	} `json:"type_config"`
	Workspace struct {
		SkillIDs []int64 `json:"skill_ids"`
	} `json:"workspace"`
}

func loadExpertProjectionRuntime(
	tx *gorm.DB,
	organizationID int64,
	snapshotID int64,
) (expertProjectionRuntime, error) {
	var row struct {
		SpecJSON json.RawMessage `gorm:"column:spec_json"`
	}
	if err := tx.Table("worker_spec_snapshots").
		Select("spec_json").
		Where("organization_id = ? AND id = ?", organizationID, snapshotID).
		First(&row).Error; err != nil || len(row.SpecJSON) == 0 {
		return expertProjectionRuntime{}, control.ErrCorrupt
	}
	var spec expertSnapshotSpec
	if err := json.Unmarshal(row.SpecJSON, &spec); err != nil {
		return expertProjectionRuntime{}, control.ErrCorrupt
	}
	if spec.Runtime.WorkerType.Slug == "" ||
		spec.TypeConfig.InteractionMode == "" ||
		spec.TypeConfig.AutomationLevel == "" {
		return expertProjectionRuntime{}, control.ErrCorrupt
	}
	slugs, err := loadExpertSkillSlugs(
		tx,
		organizationID,
		spec.Workspace.SkillIDs,
	)
	if err != nil {
		return expertProjectionRuntime{}, err
	}
	return expertProjectionRuntime{
		AgentSlug:       spec.Runtime.WorkerType.Slug,
		InteractionMode: spec.TypeConfig.InteractionMode,
		AutomationLevel: spec.TypeConfig.AutomationLevel,
		SkillSlugs:      slugs,
	}, nil
}

func loadExpertSkillSlugs(
	tx *gorm.DB,
	organizationID int64,
	skillIDs []int64,
) (pq.StringArray, error) {
	if len(skillIDs) == 0 {
		return pq.StringArray{}, nil
	}
	var rows []struct {
		ID   int64
		Slug string
	}
	if err := tx.Table("skills").
		Select("id, slug").
		Where("id IN ? AND (organization_id IS NULL OR organization_id = ?)", skillIDs, organizationID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	byID := make(map[int64]string, len(rows))
	for _, row := range rows {
		byID[row.ID] = row.Slug
	}
	slugs := make(pq.StringArray, 0, len(skillIDs))
	for _, id := range skillIDs {
		slug := byID[id]
		if slug == "" {
			return nil, control.ErrCorrupt
		}
		slugs = append(slugs, slug)
	}
	return slugs, nil
}
