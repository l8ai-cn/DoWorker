package workercreation

import (
	"context"
	"regexp"
	"strings"
	"testing"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
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
	definition := workerDefinition("codex-cli", "codex", source, "pty", "acp")
	resolver := newWorkerTypeResolver(
		provider,
		staticWorkerDefinitions{"codex-cli": definition},
	)

	resolved, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
	)

	require.NoError(t, err)
	assert.Equal(t, slugkit.MustNewForTest("codex-cli"), resolved.WorkerType.Slug)
	assert.Regexp(t, regexp.MustCompile(`^[a-f0-9]{64}$`), resolved.WorkerType.DefinitionHash)
	assert.Equal(t, definition.DefinitionHash, resolved.WorkerType.DefinitionHash)
	assert.Equal(t, uint32(1), resolved.TypeSchema.Version)
	assert.Equal(t, specdomain.TypeFieldSchema{
		Kind:        specdomain.TypeFieldSelect,
		Options:     []string{"untrusted", "on-request", "never"},
		Default:     "on-request",
		Description: "用于调整此 Agent 的运行行为。",
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

}

func TestWorkerTypeResolverExcludesModelResourceManagedFields(t *testing.T) {
	source := "AGENT worker\nEXECUTABLE cursor\n" +
		"CONFIG model STRING = \"\"\n" +
		"ENV OPENAI_API_KEY SECRET OPTIONAL\n" +
		"ENV CURSOR_API_KEY SECRET OPTIONAL\n"
	definition := workerDefinition("cursor-cli", "cursor", source, "pty", "acp")
	definition.ModelRequirement = workerdefinition.ModelRequirement{
		Required:         true,
		ProtocolAdapters: []string{"openai-compatible"},
	}
	definition.CredentialBindings = []workerdefinition.CredentialBinding{
		{
			ID: "openai",
			Source: workerdefinition.CredentialSource{
				Kind: "model_resource",
				Ref:  "openai-compatible",
			},
			Target: workerdefinition.CredentialTarget{
				Kind: "env",
				Name: "OPENAI_API_KEY",
			},
		},
		{
			ID: "cursor",
			Source: workerdefinition.CredentialSource{
				Kind: "credential_bundle",
				Ref:  "cursor",
			},
			Target: workerdefinition.CredentialTarget{
				Kind: "env",
				Name: "CURSOR_API_KEY",
			},
		},
	}
	resolver := newWorkerTypeResolver(&workerTypeAgentProvider{
		agent: activeWorkerTypeAgentFor("cursor-cli", "cursor", source),
	}, staticWorkerDefinitions{"cursor-cli": definition})

	resolved, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("cursor-cli"),
	)

	require.NoError(t, err)
	assert.NotContains(t, resolved.TypeSchema.Fields, "model")
	assert.NotContains(t, resolved.TypeSchema.Fields, "OPENAI_API_KEY")
	assert.Equal(
		t,
		specdomain.TypeFieldSecret,
		resolved.TypeSchema.Fields["CURSOR_API_KEY"].Kind,
	)
}

func TestWorkerTypeResolverProjectsCredentialRequirementGroups(t *testing.T) {
	source := "AGENT aider\nEXECUTABLE aider\n" +
		"ENV OPENAI_API_KEY SECRET OPTIONAL\n" +
		"ENV ANTHROPIC_API_KEY SECRET OPTIONAL\n"
	definition := workerDefinition("aider", "aider", source, "pty")
	definition.ModelRequirement = workerdefinition.ModelRequirement{}
	definition.CredentialBindings = []workerdefinition.CredentialBinding{
		credentialBinding("openai", "aider", "OPENAI_API_KEY"),
		credentialBinding("anthropic", "aider", "ANTHROPIC_API_KEY"),
	}
	definition.CredentialRequirementGroups = []workerdefinition.CredentialRequirementGroup{{
		ID: "provider-api-key", AnyOf: []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY"},
	}}
	provider := &workerTypeAgentProvider{
		agent: activeWorkerTypeAgentFor("aider", "aider", source),
	}
	provider.agent.SupportedModes = "pty"
	resolver := newWorkerTypeResolver(
		provider,
		staticWorkerDefinitions{"aider": definition},
	)

	resolved, err := resolver.ResolveWorkerType(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("aider"),
	)

	require.NoError(t, err)
	assert.Equal(t, definition.CredentialRequirementGroups[0].AnyOf,
		resolved.TypeSchema.SecretRequirementGroups[0].AnyOf)
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
			}, staticWorkerDefinitions{
				test.slug: workerDefinition(test.slug, "codex", source, "pty", "acp"),
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
		AdapterID:       "test-adapter",
		AgentfileSource: &source,
		IsActive:        true,
		SupportedModes:  "pty,acp",
	}
}

func workerDefinition(
	slug, executable, source string,
	modes ...string,
) workerdefinition.Definition {
	return workerdefinition.Definition{
		Slug:           slug,
		Version:        "1",
		Executable:     executable,
		AdapterID:      "test-adapter",
		DefinitionHash: strings.Repeat("a", 64),
		AgentFile:      source,
		Modes:          modes,
		ModelRequirement: workerdefinition.ModelRequirement{
			Required:         true,
			ProtocolAdapters: []string{"openai-compatible"},
		},
		CredentialBindings: []workerdefinition.CredentialBinding{
			{
				ID: "openai",
				Source: workerdefinition.CredentialSource{
					Kind: "model_resource",
					Ref:  "openai-compatible",
				},
				Target: workerdefinition.CredentialTarget{
					Kind: "env",
					Name: "OPENAI_API_KEY",
				},
			},
			{
				ID: "signing",
				Source: workerdefinition.CredentialSource{
					Kind: "credential_bundle",
					Ref:  slug,
				},
				Target: workerdefinition.CredentialTarget{
					Kind: "env",
					Name: "SIGNING_KEY",
				},
			},
		},
	}
}

func credentialBinding(
	id, ref, target string,
) workerdefinition.CredentialBinding {
	return workerdefinition.CredentialBinding{
		ID: id,
		Source: workerdefinition.CredentialSource{
			Kind: "credential_bundle", Ref: ref,
		},
		Target: workerdefinition.CredentialTarget{Kind: "env", Name: target},
	}
}
