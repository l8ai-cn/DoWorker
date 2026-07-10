package airesource

import (
	"context"

	"gorm.io/gorm"
)

type migrationParityTarget struct {
	OwnerScope           string
	OwnerID              int64
	ConnectionIdentifier string
	ResourceIdentifier   string
	ProviderKey          string
	ConnectionName       string
	ResourceName         string
	BaseURL              string
	ConfiguredFields     migrationStringList
	CredentialsEncrypted string
	ModelID              string
	ConnectionStatus     string
	ResourceStatus       string
	ConnectionEnabled    bool
	ResourceEnabled      bool
	Modalities           migrationStringList
	Capabilities         migrationStringList
	IsDefault            bool
}

func migrationParityTargetFor(
	ctx context.Context,
	tx *gorm.DB,
	kind string,
	id int64,
) (*migrationParityTarget, error) {
	var row migrationParityTarget
	result := tx.WithContext(ctx).Raw(
		`SELECT connection.owner_scope,
		        connection.owner_id,
		        connection.identifier AS connection_identifier,
		        resource.identifier AS resource_identifier,
		        connection.provider_key,
		        connection.name AS connection_name,
		        resource.display_name AS resource_name,
		        connection.base_url,
		        connection.configured_fields,
		        connection.credentials_encrypted,
		        resource.model_id,
		        connection.status AS connection_status,
		        resource.status AS resource_status,
		        connection.is_enabled AS connection_enabled,
		        resource.is_enabled AS resource_enabled,
		        resource.modalities,
		        resource.capabilities,
		        EXISTS (
		          SELECT 1
		            FROM model_resource_defaults defaults
		           WHERE defaults.model_resource_id = resource.id
		             AND defaults.owner_scope = connection.owner_scope
		             AND defaults.owner_id = connection.owner_id
		             AND defaults.modality = 'chat'
		        ) AS is_default
		   FROM ai_resource_migration_map migration
		   JOIN provider_connections connection ON connection.id = migration.provider_connection_id
		   JOIN model_resources resource ON resource.id = migration.model_resource_id
		  WHERE migration.source_kind = ? AND migration.source_id = ? AND migration.status = 'migrated'
		  LIMIT 1`,
		kind, id,
	).Scan(&row)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	return &row, nil
}
