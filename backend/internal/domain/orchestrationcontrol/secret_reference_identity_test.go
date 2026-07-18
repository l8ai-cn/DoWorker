package orchestrationcontrol

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretGuardAllowsReferenceIdentityUnderSensitiveField(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{
			"secretRefs": {
				"api-token": {
					"kind": "EnvironmentBundle",
					"name": "codex-secrets"
				}
			}
		}`),
		json.RawMessage(`{
			"secret_refs": {
				"api-token": {
					"kind": "env-bundle",
					"id": 105
				}
			}
		}`),
	}
	for _, document := range tests {
		require.NoError(t, rejectRawSecretJSON(document))
	}
}

func TestSecretGuardRejectsValueObjectUnderSensitiveField(t *testing.T) {
	err := rejectRawSecretJSON(json.RawMessage(`{
		"secretRefs": {
			"api-token": {
				"value": "not-a-reference"
			}
		}
	}`))

	require.Error(t, err)
}
