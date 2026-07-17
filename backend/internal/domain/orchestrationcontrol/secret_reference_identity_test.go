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

func TestSecretGuardAllowsToolModelEnvironmentTargets(t *testing.T) {
	require.NoError(t, rejectRawSecretJSON(json.RawMessage(`{
		"environment": {
			"api_key": "SEEDANCE_API_KEY",
			"base_url": "SEEDANCE_BASE_URL",
			"model_id": "SEEDANCE_MODEL"
		}
	}`)))
}

func TestSecretGuardRejectsInvalidToolModelEnvironmentTargets(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{
			"environment": {
				"api_key": "AKIAIOSFODNN7EXAMPLE",
				"base_url": "SEEDANCE_BASE_URL",
				"model_id": "SEEDANCE_MODEL"
			}
		}`),
		json.RawMessage(`{
			"environment": {
				"api_key": "SEEDANCE_API_KEY",
				"base_url": "SEEDANCE_BASE_URL",
				"model_id": "SEEDANCE_MODEL",
				"token": "plaintext-value"
			}
		}`),
	}
	for _, document := range tests {
		require.Error(t, rejectRawSecretJSON(document))
	}
}
