package coordinator

import "time"

// TicketExternalLink dedupes external issues against the AgentsMesh tickets they
// were synced into. The UNIQUE(org, platform, external_id) constraint is the
// idempotency key that keeps repeated scans from creating duplicate tickets.
type TicketExternalLink struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;uniqueIndex:idx_ticket_external_links_unique" json:"organization_id"`
	TicketID       int64 `gorm:"not null;index" json:"ticket_id"`

	PlatformType string `gorm:"size:32;not null;uniqueIndex:idx_ticket_external_links_unique" json:"platform_type"`
	SourceID     string `gorm:"size:255" json:"source_id"`
	ExternalID   string `gorm:"size:255;not null;uniqueIndex:idx_ticket_external_links_unique" json:"external_id"`
	ExternalURL  string `gorm:"size:1000" json:"external_url"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (TicketExternalLink) TableName() string { return "ticket_external_links" }
