package infra

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type jsonStringList []string

func (list *jsonStringList) Scan(value any) error {
	if value == nil {
		*list = nil
		return nil
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("cannot scan JSON string list from %T", value)
	}
	return json.Unmarshal(raw, list)
}

func (list jsonStringList) Value() (driver.Value, error) {
	if list == nil {
		return "[]", nil
	}
	raw, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

type providerConnectionRow struct {
	ID                   int64
	OwnerScope           airesource.OwnerScope
	OwnerID              int64
	Identifier           slugkit.Slug
	ProviderKey          slugkit.Slug
	Name                 string
	BaseURL              string
	CredentialsEncrypted string
	ConfiguredFields     jsonStringList
	Status               airesource.ConnectionStatus
	IsEnabled            bool
	LastValidatedAt      *time.Time
	ValidationError      string
	Revision             int64
	CreatedBy            int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (providerConnectionRow) TableName() string { return "provider_connections" }

func connectionRow(connection *airesource.Connection) *providerConnectionRow {
	revision := connection.Revision
	if revision <= 0 {
		revision = 1
	}
	return &providerConnectionRow{
		ID: connection.ID, OwnerScope: connection.OwnerScope, OwnerID: connection.OwnerID,
		Identifier: connection.Identifier, ProviderKey: connection.ProviderKey, Name: connection.Name,
		BaseURL: connection.BaseURL, CredentialsEncrypted: connection.CredentialsEncrypted,
		ConfiguredFields: jsonStringList(connection.ConfiguredFields), Status: connection.Status,
		IsEnabled: connection.IsEnabled, LastValidatedAt: connection.LastValidatedAt,
		ValidationError: connection.ValidationError, Revision: revision, CreatedBy: connection.CreatedBy,
		CreatedAt: connection.CreatedAt, UpdatedAt: connection.UpdatedAt,
	}
}

func (row *providerConnectionRow) domain() *airesource.Connection {
	return &airesource.Connection{
		ID: row.ID, OwnerScope: row.OwnerScope, OwnerID: row.OwnerID, Identifier: row.Identifier,
		ProviderKey: row.ProviderKey, Name: row.Name, BaseURL: row.BaseURL,
		CredentialsEncrypted: row.CredentialsEncrypted,
		ConfiguredFields:     append([]string(nil), row.ConfiguredFields...), Status: row.Status,
		IsEnabled: row.IsEnabled, LastValidatedAt: row.LastValidatedAt,
		ValidationError: row.ValidationError, Revision: row.Revision, CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

type modelResourceRow struct {
	ID                   int64
	ProviderConnectionID int64
	Identifier           slugkit.Slug
	ModelID              string
	DisplayName          string
	Modalities           jsonStringList
	Capabilities         jsonStringList
	Status               airesource.ConnectionStatus
	IsEnabled            bool
	LastValidatedAt      *time.Time
	ValidationError      string
	Revision             int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (modelResourceRow) TableName() string { return "model_resources" }

func resourceRow(resource *airesource.ModelResource) *modelResourceRow {
	revision := resource.Revision
	if revision <= 0 {
		revision = 1
	}
	modalities := make(jsonStringList, len(resource.Modalities))
	for index, modality := range resource.Modalities {
		modalities[index] = string(modality)
	}
	capabilities := make(jsonStringList, len(resource.Capabilities))
	for index, capability := range resource.Capabilities {
		capabilities[index] = string(capability)
	}
	return &modelResourceRow{
		ID: resource.ID, ProviderConnectionID: resource.ProviderConnectionID,
		Identifier: resource.Identifier, ModelID: resource.ModelID, DisplayName: resource.DisplayName,
		Modalities: modalities, Capabilities: capabilities, Status: resource.Status,
		IsEnabled: resource.IsEnabled, LastValidatedAt: resource.LastValidatedAt,
		ValidationError: resource.ValidationError, Revision: revision, CreatedAt: resource.CreatedAt, UpdatedAt: resource.UpdatedAt,
	}
}

func (row *modelResourceRow) domain() *airesource.ModelResource {
	modalities := make([]airesource.Modality, len(row.Modalities))
	for index, modality := range row.Modalities {
		modalities[index] = airesource.Modality(modality)
	}
	capabilities := make([]airesource.Capability, len(row.Capabilities))
	for index, capability := range row.Capabilities {
		capabilities[index] = airesource.Capability(capability)
	}
	return &airesource.ModelResource{
		ID: row.ID, ProviderConnectionID: row.ProviderConnectionID, Identifier: row.Identifier,
		ModelID: row.ModelID, DisplayName: row.DisplayName, Modalities: modalities,
		Capabilities: capabilities, Status: row.Status, IsEnabled: row.IsEnabled,
		LastValidatedAt: row.LastValidatedAt, ValidationError: row.ValidationError, Revision: row.Revision,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

type modelResourceDefaultRow struct {
	OwnerScope      airesource.OwnerScope
	OwnerID         int64
	Modality        airesource.Modality
	ModelResourceID int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (modelResourceDefaultRow) TableName() string { return "model_resource_defaults" }
