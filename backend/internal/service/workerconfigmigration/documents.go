package workerconfigmigration

import (
	"encoding/json"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	workerspec "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func (m *Migrator) migrateSnapshot(
	record snapshotRecord,
) (snapshotUpdate, bool, error) {
	document, err := decodeObject(record.SpecJSON)
	if err != nil {
		return snapshotUpdate{}, false, err
	}
	slug, err := requiredString(document, "runtime", "worker_type", "slug")
	if err != nil {
		return snapshotUpdate{}, false, err
	}
	changed, err := replaceLegacyBindings(
		document, slug, "config_bundle_ids", "config_document_bindings",
		m.definitions, "workspace",
	)
	if err != nil {
		return snapshotUpdate{}, false, err
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return snapshotUpdate{}, false, err
	}
	spec, err := workerspec.DecodeSpec(raw)
	if err != nil {
		return snapshotUpdate{}, false, fmt.Errorf("decode migrated spec: %w", err)
	}
	if err := validateSpecDocuments(spec, m.definitions); err != nil {
		return snapshotUpdate{}, false, err
	}
	canonical, err := workerspec.EncodeSpec(spec)
	if err != nil {
		return snapshotUpdate{}, false, err
	}
	if err := validateSnapshotSummary(record.SummaryJSON, spec); err != nil {
		return snapshotUpdate{}, false, err
	}
	return snapshotUpdate{id: record.ID, specJSON: canonical}, changed, nil
}

func (m *Migrator) migrateRevision(
	record revisionRecord,
) (revisionUpdate, bool, error) {
	document, err := decodeObject(record.CanonicalManifest)
	if err != nil {
		return revisionUpdate{}, false, err
	}
	slug, err := requiredString(document, "spec", "workerType")
	if err != nil {
		return revisionUpdate{}, false, err
	}
	changed, err := replaceLegacyBindings(
		document, slug, "configBundleRefs", "configDocumentBindings",
		m.definitions, "spec", "workspace",
	)
	if err != nil {
		return revisionUpdate{}, false, err
	}
	manifest, err := control.CanonicalJSONObject(document)
	if err != nil {
		return revisionUpdate{}, false, err
	}
	var decoded resource.Manifest
	if err := json.Unmarshal(manifest, &decoded); err != nil {
		return revisionUpdate{}, false, err
	}
	typed, err := m.registry.DecodeAndValidate(decoded)
	if err != nil {
		return revisionUpdate{}, false, fmt.Errorf("validate migrated manifest: %w", err)
	}
	worker, ok := typed.(*resource.WorkerTemplateSpec)
	if !ok {
		return revisionUpdate{}, false, fmt.Errorf("manifest is not a WorkerTemplate")
	}
	if err := validateTemplateDocuments(*worker, m.definitions); err != nil {
		return revisionUpdate{}, false, err
	}
	spec, err := control.CanonicalJSONObject(decoded.Spec)
	if err != nil {
		return revisionUpdate{}, false, err
	}
	digest, err := control.DigestCanonicalJSON(manifest)
	if err != nil {
		return revisionUpdate{}, false, err
	}
	if !changed && digest != record.Digest {
		return revisionUpdate{}, false, fmt.Errorf("modern revision digest does not match")
	}
	return revisionUpdate{
		id: record.ID, manifest: manifest, spec: spec, digest: digest,
	}, changed, nil
}

func replaceLegacyBindings(
	document map[string]any,
	slug, legacyField, modernField string,
	definitions DefinitionCatalog,
	workspacePath ...string,
) (bool, error) {
	workspace, err := requiredObject(document, workspacePath...)
	if err != nil {
		return false, err
	}
	required, err := definitionDocuments(slug, definitions)
	if err != nil {
		return false, err
	}
	legacy, legacyFound := workspace[legacyField]
	modern, modernFound := workspace[modernField]
	if legacyFound == modernFound {
		return false, fmt.Errorf("workspace must contain exactly one config binding field")
	}
	if modernFound {
		documents, bindingErr := bindingDocumentIDs(modern, modernField)
		if bindingErr != nil {
			return false, bindingErr
		}
		return false, validateDocumentSet(required, documents)
	}
	items, err := requiredSlice(legacy, legacyField)
	if err != nil {
		return false, err
	}
	bindings, err := legacyBindings(required, items, legacyField)
	if err != nil {
		return false, err
	}
	workspace[modernField] = bindings
	delete(workspace, legacyField)
	return true, nil
}
