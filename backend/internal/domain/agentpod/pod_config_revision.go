package agentpod

import (
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
)

const (
	ConfigRevisionStatusDraft    = "draft"
	ConfigRevisionStatusApplying = "applying"
	ConfigRevisionStatusActive   = "active"
	ConfigRevisionStatusFailed   = "failed"
)

type PodConfigRevision struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	PodID           int64           `gorm:"not null;uniqueIndex:uq_pod_config_revision" json:"pod_id"`
	Revision        int64           `gorm:"not null;uniqueIndex:uq_pod_config_revision" json:"revision"`
	AgentfileLayer  string          `gorm:"type:text;not null;default:''" json:"agentfile_layer"`
	Status          string          `gorm:"size:20;not null;default:'draft';index" json:"status"`
	ConfigSummary   json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"config_summary"`
	ModelResourceID *int64          `gorm:"column:model_resource_id" json:"model_resource_id,omitempty"`
	PreviewPort     int             `gorm:"column:preview_port;not null;default:0" json:"preview_port"`
	PreviewPath     string          `gorm:"column:preview_path;size:255;not null;default:'/'" json:"preview_path"`
	CreatedByID     int64           `gorm:"not null" json:"created_by_id"`
	ErrorMessage    *string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt       time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"not null;default:now()" json:"updated_at"`
	AppliedAt       *time.Time      `json:"applied_at,omitempty"`

	Pod       *Pod       `gorm:"foreignKey:PodID" json:"-"`
	CreatedBy *user.User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

func (PodConfigRevision) TableName() string {
	return "pod_config_revisions"
}
