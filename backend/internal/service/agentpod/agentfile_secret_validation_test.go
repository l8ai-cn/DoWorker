package agentpod

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAgentfileLayerSecretsAllowsModelAndBaseURL(t *testing.T) {
	layer := `CONFIG model = "claude-3-7-sonnet-20250219"
CONFIG base_url = "https://api.anthropic.com/v1"
CONFIG backup_model = "gpt-4.1-mini-2025-04-14"`

	require.NoError(t, validateAgentfileLayerSecrets(layer))
}

func TestValidateAgentfileLayerSecretsRejectsPlaintextCredentials(t *testing.T) {
	for _, tc := range []struct {
		name  string
		layer string
	}{
		{
			name:  "sensitive config key",
			layer: `CONFIG api_key = "sk-ant-api03-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		},
		{
			name:  "sensitive env key",
			layer: `ENV ANTHROPIC_API_KEY = "sk-ant-api03-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		},
		{
			name:  "credential-shaped value under ordinary key",
			layer: `CONFIG model = "sk-ant-api03-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, validateAgentfileLayerSecrets(tc.layer), ErrAgentfileSecretLiteral)
		})
	}
}
