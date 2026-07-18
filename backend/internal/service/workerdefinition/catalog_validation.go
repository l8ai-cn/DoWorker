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
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
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
	if err := validateConfigDocumentSchema(schema.Properties["config_documents"]); err != nil {
		return err
	}
	return nil
}

func validateConfigDocumentSchema(raw json.RawMessage) error {
	var documents struct {
		Type  string          `json:"type"`
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &documents); err != nil {
		return fmt.Errorf("decode config document schema: %w", err)
	}
	if documents.Type != "array" {
		return fmt.Errorf("config document schema must describe an array")
	}
	var item struct {
		Type       string                     `json:"type"`
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(documents.Items, &item); err != nil {
		return fmt.Errorf("decode config document item schema: %w", err)
	}
	if item.Type != "object" {
		return fmt.Errorf("config document item schema must describe an object")
	}
	for _, field := range []string{"id", "format", "target_path", "required"} {
		if !containsSchemaField(item.Required, field) {
			return fmt.Errorf("config document item schema must require %q", field)
		}
	}
	var format struct {
		Const string `json:"const"`
	}
	if err := json.Unmarshal(item.Properties["format"], &format); err != nil ||
		format.Const != "json" {
		return fmt.Errorf("config document schema format must be json")
	}
	var required struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(item.Properties["required"], &required); err != nil ||
		required.Type != "boolean" {
		return fmt.Errorf("config document schema required must be boolean")
	}
	return nil
}

func containsSchemaField(fields []string, wanted string) bool {
	for _, field := range fields {
		if field == wanted {
			return true
		}
	}
	return false
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
