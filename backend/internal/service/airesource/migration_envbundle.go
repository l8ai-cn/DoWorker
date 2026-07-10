package airesource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"gorm.io/gorm"
)

type bundleProviderSpec struct {
	Provider   string
	Key        string
	Model      string
	BaseURL    string
	AgentSlugs []string
}

var bundleProviderSpecs = []bundleProviderSpec{
	{"anthropic", "ANTHROPIC_API_KEY", "ANTHROPIC_MODEL", "ANTHROPIC_BASE_URL", []string{"claude", "claude-code", "do-agent"}},
	{"openai", "OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_BASE_URL", []string{"aider", "codex", "codex-cli", "cursor-cli", "do-agent", "opencode"}},
	{"gemini", "GEMINI_API_KEY", "GEMINI_MODEL", "GEMINI_BASE_URL", []string{"do-agent", "gemini", "gemini-cli"}},
	{"minimax", "MINIMAX_API_KEY", "MINIMAX_MODEL", "MINIMAX_BASE_URL", []string{"do-agent"}},
}

func (m *LegacyMigrator) migrateCredentialBundles(ctx context.Context, tx *gorm.DB, report *MigrationReport) error {
	if !tx.Migrator().HasTable("env_bundles") {
		return nil
	}
	var rows []legacyEnvBundleRow
	if err := tx.WithContext(ctx).Table("env_bundles").
		Where("kind = ? AND is_active = ?", envbundle.KindCredential, true).
		Order("id").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if migrated(ctx, tx, "env_bundle", row.ID) {
			continue
		}
		if err := m.createFromCredentialBundle(ctx, tx, row); err != nil {
			return err
		}
		report.EnvBundlesMigrated++
	}
	return nil
}

func (m *LegacyMigrator) createFromCredentialBundle(ctx context.Context, tx *gorm.DB, row legacyEnvBundleRow) error {
	plain, err := decryptBundleValues(m.cipher, row.Data)
	if err != nil {
		return fmt.Errorf("env_bundle %d credentials: %w", row.ID, err)
	}
	spec, err := inferBundleProvider(plain)
	if err != nil {
		return fmt.Errorf("env_bundle %d: %w", row.ID, err)
	}
	if err := validateBundleAgent(row.AgentSlug, spec); err != nil {
		return fmt.Errorf("env_bundle %d: %w", row.ID, err)
	}
	provider, ok := domain.Provider(spec.Provider)
	if !ok {
		return fmt.Errorf("env_bundle %d unknown provider %q", row.ID, spec.Provider)
	}
	modelID := strings.TrimSpace(plain[spec.Model])
	if modelID == "" {
		return fmt.Errorf("env_bundle %d has no model field %s", row.ID, spec.Model)
	}
	connectionID, err := m.insertConnection(ctx, tx, legacyConnectionInput{
		SourceKind: "env_bundle", SourceID: row.ID,
		OwnerScope: domain.OwnerScope(row.OwnerScope), OwnerID: row.OwnerID,
		Provider: provider, Name: row.Name, BaseURL: firstNonEmpty(plain[spec.BaseURL], provider.DefaultBaseURL),
		Credentials: map[string]string{"api_key": plain[spec.Key]}, Enabled: row.IsActive,
	})
	if err != nil {
		return err
	}
	return m.insertResource(ctx, tx, legacyResourceInput{
		ConnectionID: connectionID, SourceKind: "env_bundle", SourceID: row.ID,
		Identifier: "legacy-env-bundle-" + stringID(row.ID), Name: row.Name, ModelID: modelID,
		Enabled: row.IsActive, Default: row.KindPrimary,
		OwnerScope: domain.OwnerScope(row.OwnerScope), OwnerID: row.OwnerID,
	})
}

func decryptBundleValues(cipher Cipher, data map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(data))
	for key, encrypted := range data {
		value, err := cipher.Decrypt(encrypted)
		if err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, nil
}

func inferBundleProvider(values map[string]string) (bundleProviderSpec, error) {
	var found []bundleProviderSpec
	for _, spec := range bundleProviderSpecs {
		if strings.TrimSpace(values[spec.Key]) != "" {
			found = append(found, spec)
		}
	}
	if len(found) != 1 {
		return bundleProviderSpec{}, fmt.Errorf("expected exactly one provider API key, found %d", len(found))
	}
	return found[0], nil
}

func validateBundleAgent(agentSlug *string, spec bundleProviderSpec) error {
	if agentSlug == nil || strings.TrimSpace(*agentSlug) == "" {
		return nil
	}
	slug := strings.TrimSpace(*agentSlug)
	for _, allowed := range spec.AgentSlugs {
		if slug == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported agent %q for %s credentials", slug, spec.Provider)
}

func stringID(id int64) string { return strconv.FormatInt(id, 10) }
