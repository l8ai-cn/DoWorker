package airesource

import (
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Actor struct {
	UserID        int64
	CorrelationID string
}

type CreateConnectionInput struct {
	OwnerScope  domain.OwnerScope
	OwnerID     int64
	Identifier  slugkit.Slug
	ProviderKey slugkit.Slug
	Name        string
	BaseURL     string
	Credentials map[string]string `json:"-"`
}

type UpdateConnectionInput struct {
	Name        string
	BaseURL     string
	Credentials map[string]string `json:"-"`
}

type ConnectionView struct {
	ID               int64                   `json:"id"`
	OwnerScope       domain.OwnerScope       `json:"owner_scope"`
	OwnerID          int64                   `json:"owner_id"`
	Identifier       slugkit.Slug            `json:"identifier"`
	ProviderKey      slugkit.Slug            `json:"provider_key"`
	Name             string                  `json:"name"`
	BaseURL          string                  `json:"base_url"`
	ConfiguredFields []string                `json:"configured_fields"`
	Status           domain.ConnectionStatus `json:"status"`
	IsEnabled        bool                    `json:"is_enabled"`
	LastValidatedAt  *time.Time              `json:"last_validated_at,omitempty"`
	ValidationError  string                  `json:"validation_error,omitempty"`
	CanManage        bool                    `json:"can_manage"`
	Resources        []ResourceView          `json:"resources"`
}

type CreateResourceInput struct {
	ConnectionID int64
	Identifier   slugkit.Slug
	ModelID      string
	DisplayName  string
	Modalities   []domain.Modality
	Capabilities []domain.Capability
	IsEnabled    bool
}

type UpdateResourceInput struct {
	ModelID      string
	DisplayName  string
	Modalities   []domain.Modality
	Capabilities []domain.Capability
}

type ResourceView struct {
	ID                   int64                   `json:"id"`
	ProviderConnectionID int64                   `json:"provider_connection_id"`
	Identifier           slugkit.Slug            `json:"identifier"`
	ModelID              string                  `json:"model_id"`
	DisplayName          string                  `json:"display_name"`
	Modalities           []domain.Modality       `json:"modalities"`
	Capabilities         []domain.Capability     `json:"capabilities"`
	DefaultModalities    []domain.Modality       `json:"default_modalities"`
	Status               domain.ConnectionStatus `json:"status"`
	IsEnabled            bool                    `json:"is_enabled"`
	LastValidatedAt      *time.Time              `json:"last_validated_at,omitempty"`
	ValidationError      string                  `json:"validation_error,omitempty"`
	UsageSummary         *domain.UsageSummary    `json:"usage_summary,omitempty"`
}

type EffectiveResourceView struct {
	Connection     ConnectionView `json:"connection"`
	Resource       ResourceView   `json:"resource"`
	Selectable     bool           `json:"selectable"`
	BlockingReason BlockingReason `json:"blocking_reason,omitempty"`
}

type BlockingReason string

const (
	BlockingConnectionDisabled  BlockingReason = "connection-disabled"
	BlockingResourceDisabled    BlockingReason = "resource-disabled"
	BlockingConnectionUnchecked BlockingReason = "connection-unchecked"
	BlockingConnectionInvalid   BlockingReason = "connection-invalid"
	BlockingResourceUnchecked   BlockingReason = "resource-unchecked"
	BlockingResourceInvalid     BlockingReason = "resource-invalid"
)

type ResolutionRequirements struct {
	Modality                domain.Modality
	Capability              domain.Capability
	AllowedProtocolAdapters []string
}

type ResolvedResource struct {
	Provider    domain.ProviderDefinition
	Connection  domain.Connection
	Resource    domain.ModelResource
	Credentials map[string]string `json:"-"`
}

type ProbeInput struct {
	Provider    domain.ProviderDefinition
	BaseURL     string
	Credentials map[string]string `json:"-"`
}
