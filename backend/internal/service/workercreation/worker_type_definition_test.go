package workercreation

import (
	"context"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerTypeResolverUsesCanonicalDefinitionAndRejectsProjectionDrift(t *testing.T) {
	source := `AGENT codex
EXECUTABLE codex
CONFIG approval_mode SELECT("untrusted", "on-request", "never") = "on-request"
ENV SIGNING_KEY SECRET OPTIONAL
`
	definition := workerdefinition.Definition{
		Slug:           "codex-cli",
		Version:        "1",
		Executable:     "codex",
		AdapterID:      "codex-app-server",
		DefinitionHash: strings.Repeat("a", 64),
		AgentFile:      source,
		Modes:          []string{"pty", "acp"},
	}
	provider := &workerTypeAgentProvider{
		agent: activeWorkerTypeAgentFor("codex-cli", "codex", source),
	}
	provider.agent.AdapterID = definition.AdapterID
	resolver := newWorkerTypeResolver(provider, staticWorkerDefinitions{
		"codex-cli": definition,
	})

	resolved, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)

	require.NoError(t, err)
	assert.Equal(t, definition.DefinitionHash, resolved.WorkerType.DefinitionHash)
	assert.Contains(t, resolved.TypeSchema.Fields, "approval_mode")
	assert.Contains(t, resolved.TypeSchema.Fields, "SIGNING_KEY")

	drifted := source + "MCP ON\n"
	provider.agent.AgentfileSource = &drifted
	_, err = resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "projection")
}

func TestWorkerTypeResolverRejectsAdapterProjectionDrift(t *testing.T) {
	source := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	definition := workerDefinition("codex-cli", "codex", source, "acp")
	provider := &workerTypeAgentProvider{
		agent: activeWorkerTypeAgentFor("codex-cli", "codex", source),
	}
	provider.agent.SupportedModes = "acp"
	provider.agent.AdapterID = "different-adapter"
	resolver := newWorkerTypeResolver(provider, staticWorkerDefinitions{
		"codex-cli": definition,
	})

	_, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "projection")
}

type staticWorkerDefinitions map[string]workerdefinition.Definition

func (definitions staticWorkerDefinitions) Get(slug string) (workerdefinition.Definition, bool) {
	definition, ok := definitions[slug]
	return definition, ok
}
