package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSource_ConfigAndCredentialFields(t *testing.T) {
	src := `AGENT claude
CONFIG model SELECT("", "sonnet", "opus") = "sonnet"
CONFIG mcp_enabled BOOL = true
ENV ANTHROPIC_API_KEY SECRET OPTIONAL
ENV ANTHROPIC_AUTH_TOKEN SECRET OPTIONAL
ENV ANTHROPIC_BASE_URL TEXT OPTIONAL
ENV CODEX_HOME = sandbox.root + "/codex-home"
ENV TERM = "xterm-256color"
`
	s, err := FromSource(src)
	require.NoError(t, err)

	require.Len(t, s.ConfigFields, 2)
	assert.Equal(t, "model", s.ConfigFields[0].Name)
	assert.Equal(t, "select", s.ConfigFields[0].Type)
	assert.Equal(t, "sonnet", s.ConfigFields[0].Default)

	require.Len(t, s.CredentialFields, 3)
	assert.Equal(t, CredentialField{Name: "ANTHROPIC_API_KEY", Type: "secret", Optional: true}, s.CredentialFields[0])
	assert.Equal(t, CredentialField{Name: "ANTHROPIC_AUTH_TOKEN", Type: "secret", Optional: true}, s.CredentialFields[1])
	assert.Equal(t, CredentialField{Name: "ANTHROPIC_BASE_URL", Type: "text", Optional: true}, s.CredentialFields[2])
	assert.Equal(t, []string{"ANTHROPIC_BASE_URL"}, s.NonSecretCredentialKeys())
}

func TestFromSource_DoAgentConfigFile(t *testing.T) {
	src := `AGENT do-agent
ENV DO_AGENT_SETTINGS = sandbox.root + "/do-agent-home/settings.json"
`
	s, err := FromSource(src)
	require.NoError(t, err)
	require.Len(t, s.ConfigFiles, 1)
	assert.Equal(t, "settings", s.ConfigFiles[0].ID)
	assert.Equal(t, "DO_AGENT_SETTINGS", s.ConfigFiles[0].PathEnv)
	assert.Equal(t, "json", s.ConfigFiles[0].Format)
}

func TestFromSource_DoAgentCredentialFields(t *testing.T) {
	src := `AGENT do-agent
ENV DO_AGENT_SETTINGS = sandbox.root + "/do-agent-home/settings.json"
ENV OPENAI_API_KEY SECRET OPTIONAL
ENV ANTHROPIC_API_KEY SECRET OPTIONAL
`
	s, err := FromSource(src)
	require.NoError(t, err)
	require.Len(t, s.CredentialFields, 2)
	assert.Equal(t, "OPENAI_API_KEY", s.CredentialFields[0].Name)
	assert.Equal(t, "ANTHROPIC_API_KEY", s.CredentialFields[1].Name)
}

func TestFromSource_ParseError(t *testing.T) {
	_, err := FromSource(`INVALID @@@`)
	assert.Error(t, err)
}
