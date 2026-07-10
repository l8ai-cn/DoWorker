package airesource

import (
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type OwnerScope string

const (
	OwnerScopeUser OwnerScope = "user"
	OwnerScopeOrg  OwnerScope = "org"
)

type ConnectionStatus string

const (
	ConnectionStatusUnchecked ConnectionStatus = "unchecked"
	ConnectionStatusValid     ConnectionStatus = "valid"
	ConnectionStatusInvalid   ConnectionStatus = "invalid"
)

type Connection struct {
	ID                   int64            `json:"id"`
	OwnerScope           OwnerScope       `json:"owner_scope"`
	OwnerID              int64            `json:"owner_id"`
	Identifier           slugkit.Slug     `json:"identifier"`
	ProviderKey          slugkit.Slug     `json:"provider_key"`
	Name                 string           `json:"name"`
	BaseURL              string           `json:"base_url"`
	CredentialsEncrypted string           `json:"-"`
	ConfiguredFields     []string         `json:"configured_fields"`
	Status               ConnectionStatus `json:"status"`
	IsEnabled            bool             `json:"is_enabled"`
	LastValidatedAt      *time.Time       `json:"last_validated_at,omitempty"`
	ValidationError      string           `json:"validation_error,omitempty"`
	Revision             int64            `json:"-"`
	CreatedBy            int64            `json:"created_by"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

func (connection Connection) ValidateIdentifiers() error {
	if err := slugkit.Validate(connection.Identifier.String()); err != nil {
		return fmt.Errorf("provider connection identifier: %w", err)
	}
	if err := slugkit.Validate(connection.ProviderKey.String()); err != nil {
		return fmt.Errorf("provider connection provider key: %w", err)
	}
	return nil
}
