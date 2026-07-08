package aimodel

import "time"

// AIModel is one configured model in the org/user model pool. A Worker
// (session) is launched by referencing a model — the resolved provider
// credentials + model id are injected into the pod's do-agent settings.json.
//
// Scope: organization_id set => org-shared (visible to all org members);
// user_id set => user-private. At least one is non-null (DB CHECK). An org
// row with user_id also set is a user override of an org default.
type AIModel struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	OrganizationID *int64 `gorm:"index" json:"organization_id,omitempty"`
	UserID         *int64 `gorm:"index" json:"user_id,omitempty"`

	Name         string `gorm:"size:100;not null" json:"name"`
	ProviderType string `gorm:"size:50;not null" json:"provider_type"`
	Model        string `gorm:"size:200;not null" json:"model"`
	BaseURL      string `gorm:"size:500;not null;default:''" json:"base_url"`

	// EncryptedCredentials holds a JSON map (api_key, auth_token, ...) encrypted
	// by the service layer. Never serialized to clients.
	EncryptedCredentials string `gorm:"type:text;not null;default:''" json:"-"`

	IsDefault bool `gorm:"not null;default:false" json:"is_default"`
	IsEnabled bool `gorm:"not null;default:true" json:"is_enabled"`

	// TokenBudget is the default per-Worker token ceiling this model suggests
	// (nil = unlimited). The create-session flow may override it per Worker.
	TokenBudget *int64 `json:"token_budget,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (AIModel) TableName() string { return "ai_models" }

const (
	ProviderTypeAnthropic = "anthropic"
	ProviderTypeOpenAI    = "openai"
	ProviderTypeGemini    = "gemini"
	ProviderTypeMiniMax   = "minimax"
)
