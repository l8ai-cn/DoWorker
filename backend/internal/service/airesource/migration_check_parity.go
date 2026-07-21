package airesource

import (
	"context"
	"encoding/json"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

type expectedMigrationParity struct {
	Provider             domain.ProviderDefinition
	OwnerScope           domain.OwnerScope
	OwnerID              int64
	ConnectionIdentifier string
	ResourceIdentifier   string
	Name                 string
	BaseURL              string
	ModelID              string
	Credentials          map[string]string
	Enabled              bool
	Default              bool
}

func (m *LegacyMigrator) checkAIModelParity(
	ctx context.Context,
	tx *gorm.DB,
	row legacyAIModelRow,
	credentials map[string]string,
	report *MigrationCheckReport,
) error {
	target, err := migrationParityTargetFor(ctx, tx, "ai_model", row.ID)
	if err != nil {
		return err
	}
	if target == nil {
		report.BrokenMappings++
		return nil
	}
	provider, ok := domain.Provider(row.ProviderType)
	if !ok {
		report.FieldMismatches++
		return nil
	}
	ownerScope, ownerID, err := ownerFromLegacy(row.OrganizationID, row.UserID)
	if err != nil {
		report.ScopeMismatches++
		return nil
	}
	m.compareCommonParity(target, expectedMigrationParity{
		Provider: provider, OwnerScope: ownerScope, OwnerID: ownerID,
		ConnectionIdentifier: validIdentifier("ai_model-" + stringID(row.ID)),
		ResourceIdentifier:   validIdentifier("legacy-ai-model-" + stringID(row.ID)),
		Name:                 row.Name, BaseURL: firstNonEmpty(row.BaseURL, provider.DefaultBaseURL),
		ModelID: row.Model, Credentials: credentials,
		Enabled: row.IsEnabled, Default: row.IsDefault,
	}, report)
	return nil
}

func (m *LegacyMigrator) checkCredentialBundleParity(
	ctx context.Context,
	tx *gorm.DB,
	row legacyEnvBundleRow,
	values map[string]string,
	report *MigrationCheckReport,
) error {
	target, err := migrationParityTargetFor(ctx, tx, "env_bundle", row.ID)
	if err != nil {
		return err
	}
	if target == nil {
		report.BrokenMappings++
		return nil
	}
	spec, err := inferBundleProvider(values)
	if err != nil {
		report.FieldMismatches++
		return nil
	}
	if err := validateBundleAgent(row.AgentSlug, spec); err != nil {
		report.FieldMismatches++
		return nil
	}
	provider, ok := domain.Provider(spec.Provider)
	if !ok {
		report.FieldMismatches++
		return nil
	}
	credentials := map[string]string{"api_key": values[spec.Key]}
	m.compareCommonParity(target, expectedMigrationParity{
		Provider: provider, OwnerScope: domain.OwnerScope(row.OwnerScope), OwnerID: row.OwnerID,
		ConnectionIdentifier: validIdentifier("env_bundle-" + stringID(row.ID)),
		ResourceIdentifier:   validIdentifier("legacy-env-bundle-" + stringID(row.ID)),
		Name:                 row.Name, BaseURL: firstNonEmpty(values[spec.BaseURL], provider.DefaultBaseURL),
		ModelID: values[spec.Model], Credentials: credentials,
		Enabled: row.IsActive, Default: row.KindPrimary,
	}, report)
	return nil
}

func (m *LegacyMigrator) compareCommonParity(
	target *migrationParityTarget,
	expected expectedMigrationParity,
	report *MigrationCheckReport,
) {
	if target.OwnerScope != string(expected.OwnerScope) || target.OwnerID != expected.OwnerID {
		report.ScopeMismatches++
	}
	configured, err := validateCredentials(expected.Provider, expected.Credentials)
	if err != nil {
		report.FieldMismatches++
		return
	}
	mismatch := target.ConnectionIdentifier != expected.ConnectionIdentifier ||
		target.ResourceIdentifier != expected.ResourceIdentifier ||
		target.ProviderKey != expected.Provider.Key.String() ||
		target.ConnectionName != expected.Name ||
		target.ResourceName != expected.Name ||
		target.BaseURL != expected.BaseURL ||
		target.ModelID != expected.ModelID ||
		target.ConnectionStatus != string(domain.ConnectionStatusValid) ||
		target.ResourceStatus != string(domain.ConnectionStatusValid) ||
		target.ConnectionEnabled != expected.Enabled ||
		target.ResourceEnabled != expected.Enabled ||
		target.IsDefault != expected.Default ||
		!sameStringList(target.ConfiguredFields, configured) ||
		!sameStringList(target.Modalities, []string{string(domain.ModalityChat)}) ||
		!sameStringList(target.Capabilities, []string{string(domain.CapabilityTextGeneration)})
	actual, err := decryptedConnectionCredentials(m.cipher, target.CredentialsEncrypted)
	if err != nil {
		report.DecryptFailures++
		return
	}
	if !sameStringMap(actual, expected.Credentials) {
		mismatch = true
	}
	if mismatch {
		report.FieldMismatches++
	}
}

func decryptedConnectionCredentials(cipher Cipher, encrypted string) (map[string]string, error) {
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

func sameStringList(actual migrationStringList, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}

func sameStringMap(actual, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, expectedValue := range expected {
		if actual[key] != expectedValue {
			return false
		}
	}
	return true
}
