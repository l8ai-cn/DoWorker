package tokenquota

import "time"

const (
	PeriodTotal   = "total"
	PeriodMonthly = "monthly"
)

// TokenQuota is a token ceiling for a scope. UserID nil => org-wide;
// Model nil => applies across all models. Enforcement is report-only:
// consumption is aggregated on read and compared against LimitTokens.
type TokenQuota struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	OrganizationID int64   `gorm:"not null;index" json:"organization_id"`
	UserID         *int64  `json:"user_id,omitempty"`
	Model          *string `gorm:"size:200" json:"model,omitempty"`

	LimitTokens int64  `gorm:"not null" json:"limit_tokens"`
	Period      string `gorm:"size:20;not null;default:total" json:"period"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (TokenQuota) TableName() string { return "token_quotas" }
