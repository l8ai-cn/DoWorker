package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"maps"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

const validSubmissionJSON = `{
  "apiVersion": "agentcloud.io/v1alpha1",
  "kind": "WorkerTemplate",
  "metadata": {
    "name": "worker-one",
    "namespace": "team-one",
    "displayName": "Worker One",
    "labels": {"role": "build-agent"}
  },
  "spec": {"runtime": "codex"}
}`

func TestDecodeJSONSubmissionAcceptsValidManifest(t *testing.T) {
	manifest, err := DecodeJSONSubmission([]byte(validSubmissionJSON))

	require.NoError(t, err)
	require.Equal(t, APIVersionV1Alpha1, manifest.APIVersion)
	require.Equal(t, "WorkerTemplate", manifest.Kind)
	require.Equal(t, "worker-one", manifest.Metadata.Name.String())
	require.Equal(t, "team-one", manifest.Metadata.Namespace.String())
	require.JSONEq(t, `{"runtime":"codex"}`, string(manifest.Spec))
}

func TestDecodeJSONSubmissionRejectsOversizedSource(t *testing.T) {
	_, err := DecodeJSONSubmission(make([]byte, maxManifestBytes+1))

	require.Error(t, err)
	require.Contains(t, err.Error(), "1048576")
}

func TestDecodeJSONSubmissionRejectsInvalidUTF8(t *testing.T) {
	for _, marker := range []string{"WorkerTemplate", "codex"} {
		source := []byte(validSubmissionJSON)
		source[bytes.Index(source, []byte(marker))] = 0xff
		require.False(t, utf8.Valid(source))

		_, err := DecodeJSONSubmission(source)

		require.Error(t, err)
		require.Contains(t, err.Error(), "UTF-8")
	}
}

func TestDecodeJSONSubmissionEnforcesMaximumContainerDepth(t *testing.T) {
	manifest, err := DecodeJSONSubmission(nestedSubmissionJSON(64))
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Spec)

	_, err = DecodeJSONSubmission(nestedSubmissionJSON(65))
	require.Error(t, err)
	require.Contains(t, err.Error(), "depth")
}

func TestDecodeJSONSubmissionRejectsTrailingDocument(t *testing.T) {
	_, err := DecodeJSONSubmission([]byte(validSubmissionJSON + "\n{}"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "trailing")
}

func TestDecodeJSONSubmissionRequiresMetadataObject(t *testing.T) {
	for _, value := range []string{"null", "[]", `"metadata"`} {
		source := `{"apiVersion":"agentcloud.io/v1alpha1","kind":"WorkerTemplate",` +
			`"metadata":` + value + `,"spec":{"runtime":"codex"}}`
		_, err := DecodeJSONSubmission([]byte(source))
		require.Error(t, err)
		require.Contains(t, err.Error(), "metadata")
	}
}

func TestDecodeJSONSubmissionRejectsMissingEmptyOrNonObjectSpec(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{name: "missing", spec: ""},
		{name: "empty object", spec: `"spec": {}`},
		{name: "null", spec: `"spec": null`},
		{name: "array", spec: `"spec": []`},
		{name: "string", spec: `"spec": "codex"`},
		{name: "number", spec: `"spec": 1`},
		{name: "boolean", spec: `"spec": true`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := submissionWithFinalField(tt.spec)
			_, err := DecodeJSONSubmission([]byte(source))

			require.Error(t, err)
			require.Contains(t, err.Error(), "spec")
		})
	}
}

func TestDecodeJSONSubmissionRejectsServerOwnedFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
	}{
		{name: "uid", field: "uid", value: `"resource-123"`},
		{name: "resource version", field: "resourceVersion", value: `"42"`},
		{name: "generation", field: "generation", value: `1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := strings.Replace(
				validSubmissionJSON,
				`"name": "worker-one",`,
				`"name": "worker-one", "`+tt.field+`": `+tt.value+`,`,
				1,
			)
			_, err := DecodeJSONSubmission([]byte(source))

			require.ErrorIs(t, err, ErrServerManagedField)
			require.Contains(t, err.Error(), "metadata."+tt.field)
		})
	}

	t.Run("status", func(t *testing.T) {
		source := strings.Replace(validSubmissionJSON, `"spec":`, `"status": {"ready": true}, "spec":`, 1)
		_, err := DecodeJSONSubmission([]byte(source))

		require.ErrorIs(t, err, ErrServerManagedField)
		require.Contains(t, err.Error(), "status")
	})
}

func TestEncodeJSONProducesStableStoredManifestWithoutMutation(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Status = json.RawMessage(`{"ready":true}`)
	before := cloneManifestForJSONCodecTest(manifest)
	compact, err := json.Marshal(manifest)
	require.NoError(t, err)
	encoded, err := EncodeJSON(manifest)
	require.NoError(t, err)
	require.Equal(t, compact, encoded[:len(encoded)-1])
	require.Equal(t, byte('\n'), encoded[len(encoded)-1])
	require.NotContains(t, string(encoded[:len(encoded)-1]), "\n")
	require.Equal(t, before, manifest)
	encodedAgain, err := EncodeJSON(manifest)
	require.NoError(t, err)
	require.Equal(t, encoded, encodedAgain)
}

func TestEncodeJSONRejectsInvalidStoredManifest(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = json.RawMessage(`{}`)
	encoded, err := EncodeJSON(manifest)
	require.Error(t, err)
	require.Nil(t, encoded)
	require.Contains(t, err.Error(), "spec")
}

func TestEncodeJSONAccountsForTrailingNewlineInSizeLimit(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = json.RawMessage(`{"payload":""}`)
	base, err := json.Marshal(manifest)
	require.NoError(t, err)
	for _, target := range []int{maxManifestBytes - 1, maxManifestBytes} {
		manifest.Spec = json.RawMessage(`{"payload":"` + strings.Repeat("a", target-len(base)) + `"}`)
		compact, err := json.Marshal(manifest)
		require.NoError(t, err)
		require.Len(t, compact, target)
		encoded, err := EncodeJSON(manifest)
		if target == maxManifestBytes {
			require.Error(t, err)
			require.Nil(t, encoded)
			continue
		}
		require.NoError(t, err)
		require.Len(t, encoded, maxManifestBytes)
	}
}

func TestEncodeJSONEnforcesSafeJSONStructure(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = nestedSpecJSON(64)
	_, err := EncodeJSON(manifest)
	require.NoError(t, err)

	manifest.Spec = nestedSpecJSON(65)
	_, err = EncodeJSON(manifest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "depth")

	tests := []struct {
		name   string
		spec   json.RawMessage
		status json.RawMessage
	}{
		{name: "duplicate spec key", spec: json.RawMessage(`{"runtime":"codex","runtime":"aider"}`)},
		{name: "duplicate status key", status: json.RawMessage(`{"ready":true,"ready":false}`)},
		{name: "invalid UTF-8 spec", spec: invalidUTF8JSONObject("runtime")},
		{name: "invalid UTF-8 status", status: invalidUTF8JSONObject("message")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			if tt.spec != nil {
				manifest.Spec = tt.spec
			}
			manifest.Status = tt.status
			_, err := EncodeJSON(manifest)
			require.Error(t, err)
		})
	}
}

func nestedSubmissionJSON(depth int) []byte {
	return []byte(`{
		"apiVersion":"agentcloud.io/v1alpha1",
		"kind":"WorkerTemplate",
		"metadata":{"name":"worker-one","namespace":"team-one"},
		"spec":` + string(nestedSpecJSON(depth)) + `
	}`)
}

func nestedSpecJSON(depth int) json.RawMessage {
	value := "true"
	for currentDepth := 3; currentDepth <= depth; currentDepth++ {
		if currentDepth%2 == 0 {
			value = `{"value":` + value + `}`
			continue
		}
		value = `[` + value + `]`
	}
	return json.RawMessage(`{"payload":` + value + `}`)
}

func invalidUTF8JSONObject(field string) json.RawMessage {
	source := append([]byte(`{"`+field+`":"`), 0xff)
	return append(source, '"', '}')
}

func submissionWithFinalField(field string) string {
	if field != "" {
		field = "," + field
	}
	return `{"apiVersion":"agentcloud.io/v1alpha1","kind":"WorkerTemplate",
		"metadata":{"name":"worker-one","namespace":"team-one"}` + field + `}`
}

func cloneManifestForJSONCodecTest(manifest Manifest) Manifest {
	clone := manifest
	clone.Spec = append(json.RawMessage(nil), manifest.Spec...)
	clone.Status = append(json.RawMessage(nil), manifest.Status...)
	clone.Metadata.Labels = maps.Clone(manifest.Metadata.Labels)
	return clone
}
