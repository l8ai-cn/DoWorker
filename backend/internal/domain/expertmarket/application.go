package expertmarket

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"gorm.io/gorm"
)

type Application struct {
	ID                       int64        `gorm:"primaryKey" json:"id"`
	Slug                     slugkit.Slug `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	PublisherOrganizationID  int64        `gorm:"not null;index" json:"publisher_organization_id"`
	SourceExpertID           int64        `gorm:"not null" json:"source_expert_id"`
	PublisherUserID          int64        `gorm:"not null" json:"publisher_user_id"`
	IsOperatorOwned          bool         `gorm:"not null;default:false" json:"is_operator_owned"`
	LatestPublishedReleaseID *int64       `json:"latest_published_release_id,omitempty"`
	CreatedAt                time.Time    `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt                time.Time    `gorm:"not null;default:now()" json:"updated_at"`
}

func (Application) TableName() string {
	return "expert_market_applications"
}

func (application *Application) BeforeSave(_ *gorm.DB) error {
	return slugkit.ValidateIdentifier(
		"expert_market_applications.slug",
		string(application.Slug),
	)
}
