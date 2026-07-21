package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeJSONSubmissionRejectsDuplicateKeysAtAnyDepth(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "top level",
			source: `{
				"apiVersion":"agentcloud.io/v1alpha1",
				"kind":"WorkerTemplate",
				"kind":"RunnerTemplate",
				"metadata":{"name":"worker-one","namespace":"team-one"},
				"spec":{"runtime":"codex"}
			}`,
		},
		{
			name: "metadata",
			source: `{
				"apiVersion":"agentcloud.io/v1alpha1",
				"kind":"WorkerTemplate",
				"metadata":{"name":"worker-one","name":"worker-two","namespace":"team-one"},
				"spec":{"runtime":"codex"}
			}`,
		},
		{
			name: "spec",
			source: `{
				"apiVersion":"agentcloud.io/v1alpha1",
				"kind":"WorkerTemplate",
				"metadata":{"name":"worker-one","namespace":"team-one"},
				"spec":{"runtime":"codex","nested":{"enabled":true,"enabled":false}}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeJSONSubmission([]byte(tt.source))

			require.ErrorIs(t, err, ErrDuplicateJSONKey)
			require.Contains(t, err.Error(), "duplicate JSON key")
		})
	}
}

func TestDecodeJSONSubmissionDoesNotExposeDuplicateKey(t *testing.T) {
	for name, key := range map[string]string{
		"short": "sk_live_duplicate",
		"long":  "sk_live_long_duplicate_" + strings.Repeat("secret-", 200),
	} {
		t.Run(name, func(t *testing.T) {
			source := strings.ReplaceAll(validSubmissionJSON, `"runtime": "codex"`,
				`"`+key+`": true, "`+key+`": false`)

			_, err := DecodeJSONSubmission([]byte(source))

			require.ErrorIs(t, err, ErrDuplicateJSONKey)
			require.Contains(t, err.Error(), "duplicate JSON key")
			require.LessOrEqual(t, len(err.Error()), 512)
			requireJSONKeyNotExposed(t, err.Error(), key)
		})
	}
}

func TestDecodeJSONSubmissionDoesNotExposeUnknownFields(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		objectName string
		source     string
	}{
		{
			name:       "top level",
			key:        "sk_live_top_unknown",
			objectName: "JSON manifest",
			source: strings.Replace(
				validSubmissionJSON,
				`"spec":`,
				`"sk_live_top_unknown": true, "spec":`,
				1,
			),
		},
		{
			name:       "metadata",
			key:        "sk_live_metadata_unknown",
			objectName: "metadata",
			source: strings.Replace(
				validSubmissionJSON,
				`"displayName": "Worker One",`,
				`"sk_live_metadata_unknown": true, "displayName": "Worker One",`,
				1,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeJSONSubmission([]byte(tt.source))

			require.ErrorIs(t, err, ErrUnknownJSONField)
			require.Contains(t, err.Error(), tt.objectName)
			require.Contains(t, err.Error(), "unknown JSON field")
			require.LessOrEqual(t, len(err.Error()), 512)
			requireJSONKeyNotExposed(t, err.Error(), tt.key)
		})
	}
}

func TestDecodeJSONSubmissionRejectsNonCanonicalFieldNames(t *testing.T) {
	fields := [][2]string{
		{`"apiVersion"`, `"APIVersion"`},
		{`"kind"`, `"Kind"`},
		{`"metadata"`, `"Metadata"`},
		{`"spec"`, `"Spec"`},
		{`"name"`, `"Name"`},
		{`"namespace"`, `"Namespace"`},
		{`"displayName"`, `"DisplayName"`},
		{`"labels"`, `"Labels"`},
	}
	for _, field := range fields {
		t.Run(field[1], func(t *testing.T) {
			unknownKey := strings.Trim(field[1], `"`)
			source := strings.Replace(validSubmissionJSON, field[0], field[1], 1)

			_, err := DecodeJSONSubmission([]byte(source))

			require.ErrorIs(t, err, ErrUnknownJSONField)
			requireJSONKeyNotExposed(t, err.Error(), unknownKey)
		})
	}

	t.Run("canonical collision", func(t *testing.T) {
		source := strings.Replace(validSubmissionJSON, `"kind": "WorkerTemplate",`,
			`"kind": "WorkerTemplate", "Kind": "RunnerTemplate",`, 1)

		_, err := DecodeJSONSubmission([]byte(source))

		require.ErrorIs(t, err, ErrUnknownJSONField)
		requireJSONKeyNotExposed(t, err.Error(), "Kind")
	})
}

func TestDecodeJSONSubmissionDoesNotExposeMultipleUnknownKeys(t *testing.T) {
	longKey := "sk_live_long_unknown_" + strings.Repeat("secret-", 200)
	const otherKey = "sk_live_other_unknown"
	source := strings.Replace(validSubmissionJSON, `"spec":`,
		`"`+otherKey+`": true, "`+longKey+`": true, "spec":`, 1)

	_, err := DecodeJSONSubmission([]byte(source))

	require.ErrorIs(t, err, ErrUnknownJSONField)
	require.Contains(t, err.Error(), "JSON manifest")
	require.LessOrEqual(t, len(err.Error()), 512)
	requireJSONKeyNotExposed(t, err.Error(), longKey)
	requireJSONKeyNotExposed(t, err.Error(), otherKey)
}

func requireJSONKeyNotExposed(t *testing.T, message, key string) {
	t.Helper()
	require.NotContains(t, message, key)
	prefix := []rune(key)
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}
	require.NotContains(t, message, string(prefix))
}
