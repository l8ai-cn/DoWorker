package workerdefinition

import (
	"context"
	"errors"
	"fmt"
	"strings"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"gorm.io/gorm"
)

func SyncAgentProjections(ctx context.Context, db *gorm.DB, catalog *Catalog) (int, error) {
	if db == nil {
		return 0, errors.New("worker definition projection database is required")
	}
	if catalog == nil {
		return 0, errors.New("worker definition catalog is required")
	}
	slugs := catalog.Slugs()
	if len(slugs) == 0 {
		return 0, errors.New("worker definition catalog is empty")
	}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, slug := range slugs {
			definition, found := catalog.Get(slug)
			if !found {
				return fmt.Errorf("worker definition %q is missing from catalog", slug)
			}
			if err := syncAgentProjection(tx, definition); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, fmt.Errorf("sync worker definition projections: %w", err)
	}
	return len(slugs), nil
}

func syncAgentProjection(db *gorm.DB, definition Definition) error {
	var existing agentdomain.Agent
	err := db.Where("slug = ?", definition.Slug).Take(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(newProjectedAgent(definition)).Error
	}
	if err != nil {
		return fmt.Errorf("lookup worker definition projection %q: %w", definition.Slug, err)
	}
	return db.Model(&existing).Updates(projectedAgentFields(definition)).Error
}

func newProjectedAgent(definition Definition) *agentdomain.Agent {
	return &agentdomain.Agent{
		Slug:              definition.Slug,
		Name:              workerDisplayName(definition.Slug),
		LaunchCommand:     definition.Executable,
		Executable:        definition.Executable,
		AdapterID:         definition.AdapterID,
		AgentfileSource:   stringPointer(definition.AgentFile),
		IsBuiltin:         true,
		IsActive:          true,
		IsInternal:        definition.Internal,
		SupportedModes:    strings.Join(definition.Modes, ","),
		UsesLegacyColumns: false,
	}
}

func projectedAgentFields(definition Definition) map[string]any {
	return map[string]any{
		"launch_command":      definition.Executable,
		"executable":          definition.Executable,
		"adapter_id":          definition.AdapterID,
		"agentfile_source":    definition.AgentFile,
		"is_builtin":          true,
		"is_active":           true,
		"is_internal":         definition.Internal,
		"supported_modes":     strings.Join(definition.Modes, ","),
		"uses_legacy_columns": false,
	}
}

func stringPointer(value string) *string {
	return &value
}

func workerDisplayName(slug string) string {
	if name, found := map[string]string{
		"codex-cli":   "Codex CLI",
		"cursor-cli":  "Cursor CLI",
		"do-agent":    "Do Agent",
		"gemini-cli":  "Gemini CLI",
		"grok-build":  "Grok Build",
		"minimax-cli": "MiniMax CLI",
		"openclaw":    "OpenClaw",
		"opencode":    "OpenCode",
	}[slug]; found {
		return name
	}
	words := strings.Split(slug, "-")
	for index, word := range words {
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
