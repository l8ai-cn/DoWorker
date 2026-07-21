package airesource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"gorm.io/gorm"
)

type LegacyMigrator struct {
	db        *gorm.DB
	cipher    Cipher
	createdBy int64
}

func NewLegacyMigrator(db *gorm.DB, cipher Cipher, createdBy int64) *LegacyMigrator {
	return &LegacyMigrator{db: db, cipher: cipher, createdBy: createdBy}
}

func (m *LegacyMigrator) Run(ctx context.Context) (*MigrationReport, error) {
	if m.db == nil || m.cipher == nil || m.createdBy <= 0 {
		return nil, fmt.Errorf("AI resource migrator dependencies are incomplete")
	}
	report := &MigrationReport{}
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := m.migrateAIModels(ctx, tx, report); err != nil {
			return err
		}
		if err := syncModelResourceSequence(ctx, tx); err != nil {
			return err
		}
		if err := m.migrateCredentialBundles(ctx, tx, report); err != nil {
			return err
		}
		return m.remapVirtualKeys(ctx, tx, report)
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (m *LegacyMigrator) migrateAIModels(ctx context.Context, tx *gorm.DB, report *MigrationReport) error {
	if !tx.Migrator().HasTable("ai_models") {
		return nil
	}
	var rows []legacyAIModelRow
	if err := tx.WithContext(ctx).Table("ai_models").Order("id").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if migrated(ctx, tx, "ai_model", row.ID) {
			continue
		}
		if err := m.createFromAIModel(ctx, tx, row); err != nil {
			return err
		}
		report.AIModelsMigrated++
	}
	return nil
}

func (m *LegacyMigrator) createFromAIModel(ctx context.Context, tx *gorm.DB, row legacyAIModelRow) error {
	provider, ok := domain.Provider(row.ProviderType)
	if !ok {
		return fmt.Errorf("ai_model %d uses unknown provider %q", row.ID, row.ProviderType)
	}
	ownerScope, ownerID, err := ownerFromLegacy(row.OrganizationID, row.UserID)
	if err != nil {
		return fmt.Errorf("ai_model %d: %w", row.ID, err)
	}
	credentials, err := decryptJSONMap(m.cipher, row.EncryptedCredentials)
	if err != nil {
		return fmt.Errorf("ai_model %d credentials: %w", row.ID, err)
	}
	connectionID, err := m.insertConnection(ctx, tx, legacyConnectionInput{
		SourceKind: "ai_model", SourceID: row.ID, OwnerScope: ownerScope, OwnerID: ownerID,
		Provider: provider, Name: row.Name, BaseURL: firstNonEmpty(row.BaseURL, provider.DefaultBaseURL),
		Credentials: credentials, Enabled: row.IsEnabled,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.Model) == "" {
		return fmt.Errorf("ai_model %d has no model", row.ID)
	}
	if err := m.insertResource(ctx, tx, legacyResourceInput{
		ID: row.ID, ConnectionID: connectionID, SourceKind: "ai_model", SourceID: row.ID,
		Identifier: fmt.Sprintf("legacy-ai-model-%d", row.ID), Name: row.Name,
		ModelID: row.Model, Enabled: row.IsEnabled, Default: row.IsDefault,
		OwnerScope: ownerScope, OwnerID: ownerID,
	}); err != nil {
		return err
	}
	return nil
}

func ownerFromLegacy(orgID, userID *int64) (domain.OwnerScope, int64, error) {
	hasUser := userID != nil && *userID > 0
	hasOrg := orgID != nil && *orgID > 0
	if hasUser == hasOrg {
		return "", 0, fmt.Errorf("legacy row must have exactly one owner")
	}
	if hasUser {
		return domain.OwnerScopeUser, *userID, nil
	}
	return domain.OwnerScopeOrg, *orgID, nil
}

func migrated(ctx context.Context, tx *gorm.DB, kind string, id int64) bool {
	var count int64
	tx.WithContext(ctx).Table("ai_resource_migration_map").
		Where("source_kind = ? AND source_id = ? AND status = ?", kind, id, "migrated").
		Count(&count)
	return count > 0
}

func decryptJSONMap(cipher Cipher, encrypted string) (map[string]string, error) {
	if encrypted == "" {
		return map[string]string{}, nil
	}
	plain, err := cipher.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	if err := json.Unmarshal([]byte(plain), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func validIdentifier(value string) string {
	slug := slugkit.Sanitize(value)
	if err := slugkit.Validate(slug); err == nil {
		return slug
	}
	return "legacy-resource"
}

var _ = envbundle.KindCredential
