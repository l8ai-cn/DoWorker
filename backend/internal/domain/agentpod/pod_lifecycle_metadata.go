package agentpod

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
)

type PodLifecycleMetadata struct {
	Generation               int64      `gorm:"not null;default:0" json:"generation"`
	ActiveConfigRevisionID   *int64     `json:"active_config_revision_id,omitempty"`
	PendingConfigRevisionID  *int64     `json:"pending_config_revision_id,omitempty"`
	ReinitializeDispatchedAt *time.Time `json:"reinitialize_dispatched_at,omitempty"`
	ArchivedAt               *time.Time `json:"archived_at,omitempty"`
	ArchivedByID             *int64     `json:"archived_by_id,omitempty"`
	PurgeAfter               *time.Time `json:"purge_after,omitempty"`

	ActiveConfigRevision  *PodConfigRevision `gorm:"foreignKey:ActiveConfigRevisionID" json:"active_config_revision,omitempty"`
	PendingConfigRevision *PodConfigRevision `gorm:"foreignKey:PendingConfigRevisionID" json:"pending_config_revision,omitempty"`
	ArchivedBy            *user.User         `gorm:"foreignKey:ArchivedByID" json:"archived_by,omitempty"`
}
