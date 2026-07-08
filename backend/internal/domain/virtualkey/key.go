package virtualkey

import "time"

const (
	StatusActive    = "active"
	StatusRevoked   = "revoked"
	StatusExhausted = "exhausted"
)

// VirtualAPIKey is a platform-issued handle that wraps a real ai_models
// credential and carries an independent per-Worker token budget. Only the
// sha256 hash of the plaintext token is persisted; the token itself is
// returned once at creation. Consumption is aggregated on read from
// pod_session_usage (report-only), so no counter lives on the row.
type VirtualAPIKey struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`
	UserID         int64 `gorm:"not null;index" json:"user_id"`
	AIModelID      int64 `gorm:"not null;index" json:"ai_model_id"`

	Name      string `gorm:"size:100;not null" json:"name"`
	KeyPrefix string `gorm:"size:20;not null" json:"key_prefix"`
	KeyHash   string `gorm:"size:64;not null;uniqueIndex" json:"-"`

	TokenBudget *int64 `json:"token_budget,omitempty"`
	Status      string `gorm:"size:20;not null;default:active" json:"status"`

	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

func (VirtualAPIKey) TableName() string { return "virtual_api_keys" }
