package workerdefinition

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func validateSchema(root string) error {
	raw, err := os.ReadFile(filepath.Join(root, "schema", "definition.schema.json"))
	if err != nil {
		return fmt.Errorf("read worker definition schema: %w", err)
	}
	var schema struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return fmt.Errorf("decode worker definition schema: %w", err)
	}
	if schema.Type != "object" {
		return fmt.Errorf("worker definition schema must describe an object")
	}
	for _, field := range []string{
		"adapter_id",
		"interaction_modes",
		"model_requirement",
		"credential_bindings",
		"config_documents",
	} {
		if _, ok := schema.Properties[field]; !ok {
			return fmt.Errorf("worker definition schema is missing %q", field)
		}
	}
	return nil
}

func validateCatalogEntries(entries []catalogWorker) error {
	if len(entries) == 0 {
		return fmt.Errorf("worker definition catalog is empty")
	}
	slugs := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.Slug]; exists {
			return fmt.Errorf("worker definition catalog has duplicate slug %q", entry.Slug)
		}
		seen[entry.Slug] = struct{}{}
		expectedPath := filepath.ToSlash(filepath.Join("config", "worker-types", entry.Slug, "definition.json"))
		if entry.DefinitionPath != expectedPath {
			return fmt.Errorf("worker %q has invalid definition_path", entry.Slug)
		}
		if !strings.HasPrefix(entry.DefinitionHash, "sha256:") || len(entry.DefinitionHash) != 71 {
			return fmt.Errorf("worker %q has invalid definition_hash", entry.Slug)
		}
		slugs = append(slugs, entry.Slug)
	}
	if !sort.StringsAreSorted(slugs) {
		return fmt.Errorf("worker definition catalog slugs must be sorted")
	}
	return nil
}
