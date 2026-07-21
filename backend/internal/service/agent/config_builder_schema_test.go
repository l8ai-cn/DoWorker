package agent

import (
	"context"
	"testing"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubAgentProvider returns a pre-built agent without touching the DB. Lets
// schema tests exercise ResolveConfigSchema with any AgentFile source string,
// including malformed ones, without seeding fixtures.
type stubAgentProvider struct {
	agent *agentDomain.Agent
	err   error
}

func (s *stubAgentProvider) GetAgent(_ context.Context, _ string) (*agentDomain.Agent, error) {
	return s.agent, s.err
}

type stubCredentialFieldSource struct {
	fields map[string][]string
}

func (s stubCredentialFieldSource) CredentialBundleFields(slug string) ([]string, bool) {
	fields, found := s.fields[slug]
	return fields, found
}

var unmanagedCredentialFields = stubCredentialFieldSource{}

func TestResolveConfigSchema_ExtractsConfigFields(t *testing.T) {
	src := `AGENT claude
CONFIG model SELECT("", "sonnet", "opus") = "sonnet"
CONFIG mcp_enabled BOOL = true
CONFIG permission_mode SELECT("default", "plan", "bypassPermissions") = "bypassPermissions"
`
	p := &stubAgentProvider{
		agent: &agentDomain.Agent{Slug: "claude-code", AgentfileSource: &src},
	}

	schema, err := ResolveConfigSchema(
		context.Background(), p, unmanagedCredentialFields, "claude-code",
	)
	require.NoError(t, err)
	require.Len(t, schema.Fields, 3)

	byName := map[string]ConfigFieldResponse{}
	for _, f := range schema.Fields {
		byName[f.Name] = f
	}

	model := byName["model"]
	assert.Equal(t, "select", model.Type)
	assert.Equal(t, "sonnet", model.Default)
	require.Len(t, model.Options, 3)
	assert.Equal(t, "", model.Options[0].Value)

	mcp := byName["mcp_enabled"]
	assert.Equal(t, "boolean", mcp.Type)
	assert.Equal(t, true, mcp.Default)

	pm := byName["permission_mode"]
	assert.Equal(t, "select", pm.Type)
	assert.Equal(t, "bypassPermissions", pm.Default)
}

func TestResolveConfigSchema_ExtractsCredentialFields(t *testing.T) {
	src := `AGENT claude
ENV ANTHROPIC_API_KEY SECRET OPTIONAL
ENV ANTHROPIC_AUTH_TOKEN SECRET OPTIONAL
ENV ANTHROPIC_BASE_URL TEXT OPTIONAL
ENV CODEX_HOME = sandbox.root + "/codex-home"
`
	p := &stubAgentProvider{
		agent: &agentDomain.Agent{Slug: "claude-code", AgentfileSource: &src},
	}

	schema, err := ResolveConfigSchema(
		context.Background(), p, unmanagedCredentialFields, "claude-code",
	)
	require.NoError(t, err)
	require.Len(t, schema.CredentialFields, 3)
	assert.Equal(t, "ANTHROPIC_API_KEY", schema.CredentialFields[0].Name)
	assert.Equal(t, "secret", schema.CredentialFields[0].Type)
	assert.True(t, schema.CredentialFields[0].Optional)
	assert.Equal(t, "ANTHROPIC_BASE_URL", schema.CredentialFields[2].Name)
	assert.Equal(t, "text", schema.CredentialFields[2].Type)
}

func TestResolveConfigSchema_FiltersModelManagedCredentialFields(t *testing.T) {
	src := `AGENT cursor
ENV CURSOR_API_KEY SECRET OPTIONAL
ENV MODEL_RESOURCE_KEY SECRET OPTIONAL
`
	p := &stubAgentProvider{
		agent: &agentDomain.Agent{Slug: "cursor-cli", AgentfileSource: &src},
	}
	sources := stubCredentialFieldSource{
		fields: map[string][]string{"cursor-cli": {"CURSOR_API_KEY"}},
	}

	schema, err := ResolveConfigSchema(context.Background(), p, sources, "cursor-cli")

	require.NoError(t, err)
	require.Equal(t, []CredentialFieldResponse{{
		Name: "CURSOR_API_KEY", Type: "secret", Optional: true,
	}}, schema.CredentialFields)
}

func TestResolveConfigSchema_EmptyAgentfileReturnsEmptySchema(t *testing.T) {
	p := &stubAgentProvider{
		agent: &agentDomain.Agent{Slug: "x", AgentfileSource: nil},
	}

	schema, err := ResolveConfigSchema(
		context.Background(), p, unmanagedCredentialFields, "x",
	)
	require.NoError(t, err)
	assert.Empty(t, schema.Fields)
}

func TestResolveConfigSchema_PropagatesProviderError(t *testing.T) {
	p := &stubAgentProvider{err: assert.AnError}
	_, err := ResolveConfigSchema(
		context.Background(), p, unmanagedCredentialFields, "missing",
	)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestResolveConfigSchema_AgentFileParseError(t *testing.T) {
	src := `INVALID @@@ not real syntax`
	p := &stubAgentProvider{
		agent: &agentDomain.Agent{Slug: "x", AgentfileSource: &src},
	}

	_, err := ResolveConfigSchema(
		context.Background(), p, unmanagedCredentialFields, "x",
	)
	assert.Error(t, err, "garbage AgentFile must surface as parse error, not silent empty schema")
}
