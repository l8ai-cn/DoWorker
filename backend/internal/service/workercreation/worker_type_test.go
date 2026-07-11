package workercreation

import (
	"context"
	"regexp"
	"testing"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerTypeResolverBuildsSchemaAndStableDefinitionHash(t *testing.T) {
	source := `AGENT codex
EXECUTABLE codex
CONFIG approval_mode SELECT("untrusted", "on-request", "never") = "on-request"
CONFIG sandbox_mode BOOL = false
ENV SIGNING_KEY SECRET OPTIONAL
`
	provider := &workerTypeAgentProvider{agent: activeWorkerTypeAgent(source)}
	resolver := newWorkerTypeResolver(provider)

	resolved, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)

	require.NoError(t, err)
	assert.Equal(t, slugkit.MustNewForTest("codex-cli"), resolved.WorkerType.Slug)
	assert.Regexp(t, regexp.MustCompile(`^[a-f0-9]{64}$`), resolved.WorkerType.DefinitionHash)
	assert.Equal(t, uint32(1), resolved.TypeSchema.Version)
	assert.Equal(t, specdomain.TypeFieldSchema{
		Kind:    specdomain.TypeFieldSelect,
		Options: []string{"untrusted", "on-request", "never"},
	}, resolved.TypeSchema.Fields["approval_mode"])
	assert.Equal(t, specdomain.TypeFieldBoolean, resolved.TypeSchema.Fields["sandbox_mode"].Kind)
	assert.Equal(t, specdomain.TypeFieldSecret, resolved.TypeSchema.Fields["SIGNING_KEY"].Kind)
	assert.Equal(
		t,
		[]specdomain.InteractionMode{
			specdomain.InteractionModePTY,
			specdomain.InteractionModeACP,
		},
		resolved.SupportedInteractionModes,
	)

	again, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)
	require.NoError(t, err)
	assert.Equal(t, resolved.WorkerType.DefinitionHash, again.WorkerType.DefinitionHash)

	changed := activeWorkerTypeAgent(source + "MCP ON\n")
	provider.agent = changed
	updated, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)
	require.NoError(t, err)
	assert.NotEqual(t, resolved.WorkerType.DefinitionHash, updated.WorkerType.DefinitionHash)
}

func TestWorkerTypeResolverExcludesModelResourceManagedFields(t *testing.T) {
	tests := []struct {
		slug       string
		executable string
		source     string
		excluded   []string
	}{
		{
			slug:       "codex-cli",
			executable: "codex",
			source: "CONFIG model STRING = \"\"\n" +
				"ENV OPENAI_API_KEY SECRET OPTIONAL\n" +
				"ENV OPENAI_BASE_URL TEXT OPTIONAL\n" +
				"ENV OPENAI_MODEL TEXT OPTIONAL\n",
			excluded: []string{"model", "OPENAI_API_KEY", "OPENAI_BASE_URL", "OPENAI_MODEL"},
		},
		{
			slug:       "claude-code",
			executable: "claude",
			source: "CONFIG model STRING = \"\"\n" +
				"ENV ANTHROPIC_API_KEY SECRET OPTIONAL\n" +
				"ENV ANTHROPIC_AUTH_TOKEN SECRET OPTIONAL\n" +
				"ENV ANTHROPIC_BASE_URL TEXT OPTIONAL\n",
			excluded: []string{"model", "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL"},
		},
		{
			slug:       "gemini-cli",
			executable: "gemini",
			source: "CONFIG model STRING = \"\"\n" +
				"ENV GEMINI_API_KEY SECRET OPTIONAL\n" +
				"ENV GOOGLE_API_KEY SECRET OPTIONAL\n",
			excluded: []string{"model", "GEMINI_API_KEY", "GOOGLE_API_KEY"},
		},
	}

	for _, test := range tests {
		t.Run(test.slug, func(t *testing.T) {
			source := "AGENT worker\nEXECUTABLE " + test.executable + "\n" +
				test.source + "ENV SIGNING_KEY SECRET OPTIONAL\n"
			resolver := newWorkerTypeResolver(&workerTypeAgentProvider{
				agent: activeWorkerTypeAgentFor(test.slug, test.executable, source),
			})

			resolved, err := resolver.ResolveWorkerType(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				slugkit.MustNewForTest(test.slug),
			)

			require.NoError(t, err)
			for _, field := range test.excluded {
				assert.NotContains(t, resolved.TypeSchema.Fields, field)
			}
			assert.Equal(
				t,
				specdomain.TypeFieldSecret,
				resolved.TypeSchema.Fields["SIGNING_KEY"].Kind,
			)
		})
	}
}

func TestWorkerTypeResolverRejectsUnavailableDefinitions(t *testing.T) {
	source := "AGENT codex\n"
	tests := []struct {
		name  string
		agent *agentdomain.Agent
		err   error
		match string
		slug  string
	}{
		{
			name:  "missing worker type",
			err:   agentservice.ErrAgentNotFound,
			match: "worker type",
			slug:  "codex-cli",
		},
		{
			name: "inactive worker type",
			agent: func() *agentdomain.Agent {
				agent := activeWorkerTypeAgent(source)
				agent.IsActive = false
				return agent
			}(),
			match: "disabled",
			slug:  "codex-cli",
		},
		{
			name: "internal worker type",
			agent: func() *agentdomain.Agent {
				agent := activeWorkerTypeAgent(source)
				agent.IsInternal = true
				return agent
			}(),
			match: "internal",
			slug:  "codex-cli",
		},
		{
			name:  "provider substitutes another slug",
			agent: activeWorkerTypeAgent(source),
			match: "slug",
			slug:  "claude-code",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := newWorkerTypeResolver(&workerTypeAgentProvider{
				agent: test.agent,
				err:   test.err,
			})

			_, err := resolver.ResolveWorkerType(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				slugkit.MustNewForTest(test.slug),
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

type workerTypeAgentProvider struct {
	agent *agentdomain.Agent
	err   error
}

func (provider *workerTypeAgentProvider) GetAgent(
	context.Context,
	string,
) (*agentdomain.Agent, error) {
	return provider.agent, provider.err
}

func (provider *workerTypeAgentProvider) ListBuiltinAgents(
	context.Context,
) ([]*agentdomain.Agent, error) {
	if provider.err != nil {
		return nil, provider.err
	}
	if provider.agent == nil {
		return nil, nil
	}
	return []*agentdomain.Agent{provider.agent}, nil
}

func activeWorkerTypeAgent(source string) *agentdomain.Agent {
	return activeWorkerTypeAgentFor("codex-cli", "codex", source)
}

func activeWorkerTypeAgentFor(slug, executable, source string) *agentdomain.Agent {
	return &agentdomain.Agent{
		Slug:            slug,
		Name:            slug,
		Executable:      executable,
		AgentfileSource: &source,
		IsActive:        true,
		SupportedModes:  "pty,acp",
	}
}
