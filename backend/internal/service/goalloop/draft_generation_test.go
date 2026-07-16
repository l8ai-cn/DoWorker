package goalloop

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	airesourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

const draftLoopSource = `@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """fix checkout tax""" }
    @id(n-tests)
    verify tests { command "pnpm test" accept "tests pass" }
  }
  on_failure pause
}`

func TestDraftGeneratorResolvesExplicitModelAndCompilesSource(t *testing.T) {
	resources := &draftResourceResolver{resolved: draftResolvedResource(t)}
	raw, err := json.Marshal(loopGenerationEnvelope{Source: draftLoopSource})
	require.NoError(t, err)
	provider := &draftJSONGenerator{output: raw}
	generator := NewDraftGenerator(resources, provider)

	proposal, err := generator.Generate(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftGenerationInput{
			Prompt:          "Create a professional PPT loop",
			CurrentSource:   draftLoopSource,
			ModelResourceID: 42,
			Locale:          "zh-CN",
		},
	)

	require.NoError(t, err)
	require.NotNil(t, proposal.Program)
	require.NotEmpty(t, proposal.CanonicalSource)
	require.Equal(t, int64(7), resources.orgID)
	require.Equal(t, int64(9), resources.actor.UserID)
	require.Equal(t, int64(42), resources.resourceID)
	require.Equal(t, airesourcedomain.ModalityChat, resources.required.Modality)
	require.Equal(t, airesourcedomain.CapabilityTextGeneration, resources.required.Capability)
	require.Equal(t, supportedDraftAdapters, resources.required.AllowedProtocolAdapters)
	require.Contains(t, provider.systemPrompt, "Do not execute")
	require.Contains(t, provider.userPrompt, "professional PPT")
}

func TestDraftGeneratorRejectsMalformedOrInvalidOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{name: "unknown field", output: `{"source":"loop x {}","worker":"coder"}`},
		{name: "code fence", output: "```json\n{\"source\":\"loop x {}\"}\n```"},
		{name: "invalid LoopScript", output: `{"source":"loop broken {}"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := NewDraftGenerator(
				&draftResourceResolver{resolved: draftResolvedResource(t)},
				&draftJSONGenerator{output: []byte(test.output)},
			)

			_, err := generator.Generate(
				context.Background(),
				DraftGenerationScope{OrganizationID: 7, UserID: 9},
				DraftGenerationInput{
					Prompt: "make loop", ModelResourceID: 42, Locale: "zh-CN",
				},
			)

			require.ErrorIs(t, err, ErrGeneratedDraftInvalid)
		})
	}
}

func TestDraftGeneratorDoesNotLeakProviderFailure(t *testing.T) {
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		&draftJSONGenerator{err: errors.New("secret provider body")},
	)

	_, err := generator.Generate(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftGenerationInput{
			Prompt: "make loop", ModelResourceID: 42, Locale: "zh-CN",
		},
	)

	require.ErrorIs(t, err, ErrDraftProviderUnavailable)
	require.NotContains(t, err.Error(), "secret")
}

func TestDraftGeneratorRejectsSecretLikeInputBeforeProvider(t *testing.T) {
	provider := &draftJSONGenerator{}
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		provider,
	)

	_, err := generator.Generate(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftGenerationInput{
			Prompt:        "repair with api_key=sk-live-not-for-model",
			CurrentSource: draftLoopSource, ModelResourceID: 42, Locale: "zh-CN",
		},
	)

	require.ErrorIs(t, err, ErrDraftContainsSecret)
	require.False(t, provider.called)
}

func TestDraftGeneratorRejectsSecretLikeModelOutput(t *testing.T) {
	secretSource := strings.Replace(
		draftLoopSource,
		"fix checkout tax",
		"use api_key=sk-live-not-for-workbench",
		1,
	)
	raw, err := json.Marshal(loopGenerationEnvelope{Source: secretSource})
	require.NoError(t, err)
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		&draftJSONGenerator{output: raw},
	)

	_, err = generator.Generate(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftGenerationInput{
			Prompt: "repair loop", CurrentSource: draftLoopSource,
			ModelResourceID: 42, Locale: "zh-CN",
		},
	)

	require.ErrorIs(t, err, ErrGeneratedDraftInvalid)
}

func TestDraftGeneratorRejectsProtectedSemanticChanges(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "verification command",
			source: strings.Replace(
				draftLoopSource, `command "pnpm test"`, `command "true"`, 1,
			),
		},
		{
			name: "weaker token budget",
			source: strings.Replace(
				draftLoopSource, "tokens: 80000", "tokens: 90000", 1,
			),
		},
		{
			name: "node identity",
			source: strings.Replace(
				draftLoopSource, "n-fix-tax", "n-rewritten-task", 1,
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(loopGenerationEnvelope{Source: test.source})
			require.NoError(t, err)
			generator := NewDraftGenerator(
				&draftResourceResolver{resolved: draftResolvedResource(t)},
				&draftJSONGenerator{output: raw},
			)

			_, err = generator.Generate(
				context.Background(),
				DraftGenerationScope{OrganizationID: 7, UserID: 9},
				DraftGenerationInput{
					Prompt: "update the task", CurrentSource: draftLoopSource,
					ModelResourceID: 42, Locale: "zh-CN",
				},
			)

			require.ErrorIs(t, err, ErrGeneratedDraftInvalid)
		})
	}
}

func TestDraftGeneratorAllowsTaskChangeWithStricterLimits(t *testing.T) {
	source := strings.Replace(draftLoopSource, "fix checkout tax", "build a professional PPT", 1)
	source = strings.Replace(source, "tokens: 80000", "tokens: 70000", 1)
	raw, err := json.Marshal(loopGenerationEnvelope{Source: source})
	require.NoError(t, err)
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		&draftJSONGenerator{output: raw},
	)

	proposal, err := generator.Generate(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftGenerationInput{
			Prompt: "build a professional PPT", CurrentSource: draftLoopSource,
			ModelResourceID: 42, Locale: "zh-CN",
		},
	)

	require.NoError(t, err)
	require.Equal(t, "build a professional PPT", proposal.Program.Loop.Repeat.Agent.Prompt)
	require.Equal(t, int64(70000), proposal.Program.Loop.Limits.Tokens)
}

type draftResourceResolver struct {
	resolved   *airesourceservice.ResolvedResource
	err        error
	actor      airesourceservice.Actor
	orgID      int64
	resourceID int64
	required   airesourceservice.ResolutionRequirements
}

func (resolver *draftResourceResolver) ResolveExact(
	_ context.Context,
	actor airesourceservice.Actor,
	orgID, resourceID int64,
	required airesourceservice.ResolutionRequirements,
) (*airesourceservice.ResolvedResource, error) {
	resolver.actor = actor
	resolver.orgID = orgID
	resolver.resourceID = resourceID
	resolver.required = required
	return resolver.resolved, resolver.err
}

type draftJSONGenerator struct {
	output       []byte
	err          error
	systemPrompt string
	userPrompt   string
	called       bool
}

func (generator *draftJSONGenerator) Generate(
	_ context.Context,
	_ *airesourceservice.ResolvedResource,
	systemPrompt, userPrompt string,
) ([]byte, error) {
	generator.called = true
	generator.systemPrompt = systemPrompt
	generator.userPrompt = userPrompt
	return generator.output, generator.err
}

func draftResolvedResource(t *testing.T) *airesourceservice.ResolvedResource {
	t.Helper()
	provider, ok := airesourcedomain.Provider("openai")
	require.True(t, ok)
	return &airesourceservice.ResolvedResource{Provider: provider}
}
