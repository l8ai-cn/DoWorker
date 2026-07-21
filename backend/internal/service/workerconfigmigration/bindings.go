package workerconfigmigration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	workerspec "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func definitionDocuments(
	slug string,
	definitions DefinitionCatalog,
) ([]string, error) {
	definition, found := definitions.Get(slug)
	if !found {
		return nil, fmt.Errorf("worker type %q has no definition", slug)
	}
	documents := make([]string, len(definition.ConfigDocuments))
	seen := make(map[string]struct{}, len(documents))
	for index, document := range definition.ConfigDocuments {
		if document.ID == "" {
			return nil, fmt.Errorf("worker type %q has an empty document ID", slug)
		}
		if _, exists := seen[document.ID]; exists {
			return nil, fmt.Errorf("worker type %q repeats document %q", slug, document.ID)
		}
		seen[document.ID] = struct{}{}
		documents[index] = document.ID
	}
	sort.Strings(documents)
	return documents, nil
}

func legacyBindings(
	documents []string,
	items []any,
	field string,
) ([]any, error) {
	switch {
	case len(documents) == 0 && len(items) == 0:
		return []any{}, nil
	case len(documents) == 1 && len(items) == 1:
		if field == "config_bundle_ids" {
			id, err := positiveInt64(items[0], field)
			if err != nil {
				return nil, err
			}
			return []any{map[string]any{
				"document_id": documents[0], "config_bundle_id": id,
			}}, nil
		}
		if _, ok := items[0].(map[string]any); !ok {
			return nil, fmt.Errorf("%s must contain resource references", field)
		}
		return []any{map[string]any{
			"documentId": documents[0], "configBundleRef": items[0],
		}}, nil
	default:
		return nil, fmt.Errorf(
			"%s cannot map %d legacy binding(s) to %d definition document(s)",
			field, len(items), len(documents),
		)
	}
}

func bindingDocumentIDs(value any, field string) ([]string, error) {
	items, err := requiredSlice(value, field)
	if err != nil {
		return nil, err
	}
	documents := make([]string, len(items))
	for index, item := range items {
		binding, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s[%d] must be an object", field, index)
		}
		document, ok := binding["document_id"].(string)
		if !ok {
			document, ok = binding["documentId"].(string)
		}
		if !ok || document == "" {
			return nil, fmt.Errorf("%s[%d] has no document ID", field, index)
		}
		documents[index] = document
	}
	return documents, nil
}

func validateDocumentSet(required, actual []string) error {
	if len(required) != len(actual) {
		return fmt.Errorf("config binding count %d does not match definition count %d", len(actual), len(required))
	}
	actual = append([]string{}, actual...)
	sort.Strings(actual)
	for index := range required {
		if required[index] != actual[index] {
			return fmt.Errorf("config bindings do not match the Worker definition")
		}
	}
	return nil
}

func validateSpecDocuments(spec workerspec.Spec, definitions DefinitionCatalog) error {
	required, err := definitionDocuments(spec.Runtime.WorkerType.Slug.String(), definitions)
	if err != nil {
		return err
	}
	actual := make([]string, len(spec.Workspace.ConfigDocumentBindings))
	for index, binding := range spec.Workspace.ConfigDocumentBindings {
		actual[index] = binding.DocumentID
	}
	return validateDocumentSet(required, actual)
}

func validateTemplateDocuments(
	spec resource.WorkerTemplateSpec,
	definitions DefinitionCatalog,
) error {
	required, err := definitionDocuments(spec.WorkerType.String(), definitions)
	if err != nil {
		return err
	}
	actual := make([]string, len(spec.Workspace.ConfigDocumentBindings))
	for index, binding := range spec.Workspace.ConfigDocumentBindings {
		actual[index] = binding.DocumentID
	}
	return validateDocumentSet(required, actual)
}

func validateSnapshotSummary(raw json.RawMessage, spec workerspec.Spec) error {
	summary, err := workerspec.DecodeSummary(raw)
	if err != nil {
		return fmt.Errorf("decode snapshot summary: %w", err)
	}
	expected, err := workerspec.Summarize(spec)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(summary, expected) {
		return fmt.Errorf("snapshot summary does not match its spec")
	}
	return nil
}
