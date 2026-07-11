package airesource

import (
	"context"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

type legacyConnectionInput struct {
	SourceKind  string
	SourceID    int64
	OwnerScope  domain.OwnerScope
	OwnerID     int64
	Provider    domain.ProviderDefinition
	Name        string
	BaseURL     string
	Credentials map[string]string
	Enabled     bool
}

type legacyResourceInput struct {
	ID           int64
	ConnectionID int64
	SourceKind   string
	SourceID     int64
	Identifier   string
	Name         string
	ModelID      string
	Enabled      bool
	Default      bool
	OwnerScope   domain.OwnerScope
	OwnerID      int64
}

func (m *LegacyMigrator) insertConnection(ctx context.Context, tx *gorm.DB, in legacyConnectionInput) (int64, error) {
	encrypted, configured, err := m.encryptConnectionCredentials(in.Provider, in.Credentials)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	row := &migrationConnectionRow{
		OwnerScope: string(in.OwnerScope), OwnerID: in.OwnerID,
		Identifier:  validIdentifier(in.SourceKind + "-" + stringID(in.SourceID)),
		ProviderKey: in.Provider.Key.String(), Name: in.Name, BaseURL: in.BaseURL,
		CredentialsEncrypted: encrypted, ConfiguredFields: migrationStringList(configured),
		Status: "valid", IsEnabled: in.Enabled, CreatedBy: m.createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.WithContext(ctx).Create(row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (m *LegacyMigrator) insertResource(ctx context.Context, tx *gorm.DB, in legacyResourceInput) error {
	now := time.Now()
	row := &migrationResourceRow{
		ID: in.ID, ProviderConnectionID: in.ConnectionID, Identifier: validIdentifier(in.Identifier),
		ModelID: in.ModelID, DisplayName: in.Name,
		Modalities:   migrationStringList{string(domain.ModalityChat)},
		Capabilities: migrationStringList{string(domain.CapabilityTextGeneration)},
		Status:       "valid", IsEnabled: in.Enabled, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	resourceID := row.ID
	if in.Default {
		if err := tx.WithContext(ctx).Exec(
			`INSERT INTO model_resource_defaults(owner_scope, owner_id, modality, model_resource_id)
			 VALUES (?, ?, ?, ?)`,
			string(in.OwnerScope), in.OwnerID, string(domain.ModalityChat), resourceID,
		).Error; err != nil {
			return err
		}
	}
	return tx.WithContext(ctx).Exec(
		`INSERT INTO ai_resource_migration_map
		 (source_kind, source_id, provider_connection_id, model_resource_id, status)
		 VALUES (?, ?, ?, ?, ?)`,
		in.SourceKind, in.SourceID, in.ConnectionID, resourceID, "migrated",
	).Error
}

func syncModelResourceSequence(ctx context.Context, tx *gorm.DB) error {
	if tx.Name() != "postgres" || !tx.Migrator().HasTable("model_resources") {
		return nil
	}
	var sequenceValue int64
	return tx.WithContext(ctx).Raw(
		`SELECT setval(
			pg_get_serial_sequence('model_resources', 'id'),
			COALESCE((SELECT MAX(id) FROM model_resources), 1),
			EXISTS (SELECT 1 FROM model_resources)
		)`,
	).Scan(&sequenceValue).Error
}

func (m *LegacyMigrator) encryptConnectionCredentials(
	provider domain.ProviderDefinition,
	credentials map[string]string,
) (string, []string, error) {
	svc := &Service{cipher: m.cipher}
	return svc.encryptCredentials(provider, credentials)
}
