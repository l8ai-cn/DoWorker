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
	if len(entries) != len(formalWorkerSlugs) {
		return fmt.Errorf("worker definition catalog has %d entries, want %d", len(entries), len(formalWorkerSlugs))
	}
	slugs := make([]string, 0, len(entries))
	for _, entry := range entries {
		expectedPath := filepath.ToSlash(filepath.Join("config", "worker-types", entry.Slug, "definition.json"))
		if entry.DefinitionPath != expectedPath {
			return fmt.Errorf("worker %q has invalid definition_path", entry.Slug)
		}
		if !strings.HasPrefix(entry.DefinitionHash, "sha256:") || len(entry.DefinitionHash) != 71 {
			return fmt.Errorf("worker %q has invalid definition_hash", entry.Slug)
		}
		slugs = append(slugs, entry.Slug)
	}
	sort.Strings(slugs)
	if strings.Join(slugs, ",") != strings.Join(formalWorkerSlugs, ",") {
		return fmt.Errorf("worker definition catalog does not match formal Worker slugs")
	}
	return nil
}
