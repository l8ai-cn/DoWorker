package workerconfigmigration

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDefinitions map[string]workerdefinition.Definition

func (definitions testDefinitions) Get(slug string) (workerdefinition.Definition, bool) {
	definition, found := definitions[slug]
	return definition, found
}

func TestReplaceLegacySnapshotBinding(t *testing.T) {
	document := decodeTestObject(t, `{
		"workspace":{"config_bundle_ids":[42]}
	}`)
	changed, err := replaceLegacyBindings(
		document, "do-agent", "config_bundle_ids", "config_document_bindings",
		testDefinitions{"do-agent": definitionWithDocuments("settings")}, "workspace",
	)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.JSONEq(t, `{
		"workspace":{"config_document_bindings":[
			{"document_id":"settings","config_bundle_id":42}
		]}
	}`, encodeTestObject(t, document))
}

func TestReplaceLegacyTemplateBinding(t *testing.T) {
	document := decodeTestObject(t, `{
		"spec":{"workspace":{"configBundleRefs":[
			{"kind":"EnvironmentBundle","namespace":"dev-org","name":"settings"}
		]}}
	}`)
	changed, err := replaceLegacyBindings(
		document, "do-agent", "configBundleRefs", "configDocumentBindings",
		testDefinitions{"do-agent": definitionWithDocuments("settings")}, "spec", "workspace",
	)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.JSONEq(t, `{
		"spec":{"workspace":{"configDocumentBindings":[
			{"documentId":"settings","configBundleRef":{
				"kind":"EnvironmentBundle","namespace":"dev-org","name":"settings"
			}}
		]}}
	}`, encodeTestObject(t, document))
}

func TestReplaceLegacyBindingsRejectsAmbiguity(t *testing.T) {
	tests := []struct {
		name       string
		definition workerdefinition.Definition
		bindings   string
	}{
		{"missing required document", definitionWithDocuments("settings"), "[]"},
		{"multiple definition documents", definitionWithDocuments("one", "two"), "[1,2]"},
		{"unexpected legacy bundle", definitionWithDocuments(), "[1]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := decodeTestObject(t, `{"workspace":{"config_bundle_ids":`+test.bindings+`}}`)
			changed, err := replaceLegacyBindings(
				document, "worker", "config_bundle_ids", "config_document_bindings",
				testDefinitions{"worker": test.definition}, "workspace",
			)
			assert.False(t, changed)
			assert.Error(t, err)
		})
	}
}

func TestReplaceLegacyBindingsRejectsUndeclaredModernDocument(t *testing.T) {
	document := decodeTestObject(t, `{
		"workspace":{"config_document_bindings":[
			{"document_id":"other","config_bundle_id":1}
		]}
	}`)
	changed, err := replaceLegacyBindings(
		document, "worker", "config_bundle_ids", "config_document_bindings",
		testDefinitions{"worker": definitionWithDocuments("settings")}, "workspace",
	)
	assert.False(t, changed)
	assert.Error(t, err)
}

func definitionWithDocuments(ids ...string) workerdefinition.Definition {
	documents := make([]workerdefinition.ConfigDocument, len(ids))
	for index, id := range ids {
		documents[index] = workerdefinition.ConfigDocument{ID: id, Format: "json", TargetPath: id}
	}
	return workerdefinition.Definition{Slug: "worker", ConfigDocuments: documents}
}

func decodeTestObject(t *testing.T, source string) map[string]any {
	t.Helper()
	object, err := decodeObject([]byte(source))
	require.NoError(t, err)
	return object
}

func encodeTestObject(t *testing.T, object map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(object)
	require.NoError(t, err)
	return string(bytes.TrimSpace(raw))
}
