package imbridge

import (
	"encoding/json"
	"time"
)

const (
	ProviderFeishu   = "feishu"
	ProviderDingTalk = "dingtalk"
	ProviderWeCom    = "wecom"
	ProviderSlack    = "slack"
	ProviderWeixin   = "weixin"
	ProviderWeChat   = "wechat" // alias for weixin (OpenClaw naming)
)

const (
	StatusDisabled = "disabled"
	StatusActive   = "active"
	StatusError    = "error"
)

var SupportedProviders = []string{
	ProviderFeishu,
	ProviderDingTalk,
	ProviderWeCom,
	ProviderSlack,
	ProviderWeixin,
}

type Connection struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	OrganizationID  int64           `gorm:"not null;index" json:"organization_id"`
	Provider        string          `gorm:"size:32;not null" json:"provider"`
	Name            string          `gorm:"size:255;not null" json:"name"`
	ChannelID       *int64          `json:"channel_id,omitempty"`
	Config          json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	WebhookToken    string          `gorm:"size:64;not null" json:"-"`
	Status          string          `gorm:"size:32;not null;default:'disabled'" json:"status"`
	LastError       *string         `gorm:"type:text" json:"last_error,omitempty"`
	CreatedByUserID int64           `gorm:"not null" json:"created_by_user_id"`
	CreatedAt       time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"not null;default:now()" json:"updated_at"`

	WebhookURL string `gorm:"-" json:"webhook_url,omitempty"`
}

type ThreadMapping struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	ConnectionID     int64     `gorm:"not null;index" json:"connection_id"`
	ExternalThreadID string    `gorm:"size:512;not null" json:"external_thread_id"`
	ChannelID        int64     `gorm:"not null;index" json:"channel_id"`
	ContextToken     *string   `gorm:"size:512" json:"context_token,omitempty"`
	CreatedAt        time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Connection) TableName() string     { return "im_channel_connections" }
func (ThreadMapping) TableName() string { return "im_thread_mappings" }
