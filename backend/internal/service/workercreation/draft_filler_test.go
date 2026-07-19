package workercreation

import (
	"context"
	"errors"
	"strings"
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraftFillerAppliesAllowedPatchAndPreflights(t *testing.T) {
	fixture, generator, filler := newDraftFillerFixture(`{
		"type_config_values":{"approval_mode":"on-request"},
		"interaction_mode":"pty",
		"automation_level":"auto_edit",
		"branch":"feature/review",
		"instructions":"Review changes before editing.",
		"initial_task":"Fix the failing worker tests.",
		"termination_policy":"idle",
		"idle_timeout_minutes":45,
		"alias":"review-worker"
	}`)
	current := validWorkerCreationDraft()

	result, err := filler.Fill(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		"Create a cautious review worker",
		101,
		&current,
	)

	require.NoError(t, err)
	require.Empty(t, result.Issues)
	assert.Equal(t, "on-request", result.Draft.WorkerSpec.TypeConfig.Values["approval_mode"])
	assert.Equal(t, specdomain.InteractionModePTY, result.Draft.WorkerSpec.TypeConfig.InteractionMode)
	assert.Equal(t, specdomain.AutomationLevelAutoEdit, result.Draft.WorkerSpec.TypeConfig.AutomationLevel)
	assert.Equal(t, "feature/review", result.Draft.WorkerSpec.Workspace.Branch)
	assert.Equal(t, "Review changes before editing.", result.Draft.WorkerSpec.Workspace.Instructions)
	assert.Equal(t, "Fix the failing worker tests.", result.Draft.WorkerSpec.Workspace.InitialTask)
	assert.Equal(t, specdomain.TerminationPolicyOnIdle, result.Draft.WorkerSpec.Lifecycle.TerminationPolicy)
	assert.Equal(t, uint32(45), result.Draft.WorkerSpec.Lifecycle.IdleTimeoutMinutes)
	assert.Equal(t, "review-worker", result.Draft.WorkerSpec.Metadata.Alias)
	assert.Equal(t, int64(101), result.Draft.WorkerSpec.ModelResourceID)
	assert.Equal(t, int64(1), result.Draft.WorkerSpec.Runtime.RuntimeImageID)
	assert.Equal(t, "never", current.WorkerSpec.TypeConfig.Values["approval_mode"])
	assert.Equal(t, "main", current.WorkerSpec.Workspace.Branch)
	assert.Equal(t, 2, fixture.resources.calls)
	require.NotNil(t, generator.resource)
	assert.Equal(t, int64(101), generator.resource.Resource.ID)
	assert.NotContains(t, generator.systemPrompt, "must-not-leak")
	assert.NotContains(t, generator.userPrompt, "must-not-leak")
	assert.Contains(t, generator.userPrompt, "Create a cautious review worker")
}

func TestDraftFillerRejectsUntrustedOutputShape(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{name: "protected field", output: `{"model_resource_id":999}`},
		{name: "markdown wrapper", output: "```json\n{\"alias\":\"worker\"}\n```"},
		{name: "trailing document", output: `{"alias":"worker"} {}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, filler := newDraftFillerFixture(test.output)
			current := validWorkerCreationDraft()

			_, err := filler.Fill(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				"Fill the worker",
				101,
				&current,
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
		})
	}
}

func TestDraftFillerReturnsPreflightIssuesForInvalidPatch(t *testing.T) {
	_, _, filler := newDraftFillerFixture(`{"termination_policy":"idle"}`)
	current := validWorkerCreationDraft()

	result, err := filler.Fill(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		"Stop when idle",
		101,
		&current,
	)

	require.NoError(t, err)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "blocking", result.Issues[0].Severity)
	assert.Equal(t, specdomain.TerminationPolicyOnIdle, result.Draft.WorkerSpec.Lifecycle.TerminationPolicy)
	assert.Zero(t, result.Draft.WorkerSpec.Lifecycle.IdleTimeoutMinutes)
}

func TestDraftFillerUsesSeparateGenerationModelForWorkerWithoutMainModel(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	source := *fixture.agents.agent.AgentfileSource
	definition := noModelWorkerDefinition("codex-cli", "codex", source, "pty", "acp")
	fixture.definitions["codex-cli"] = definition
	generator := &recordingDraftJSONGenerator{output: []byte(`{"alias":"cursor-worker"}`)}
	filler := NewDraftFiller(NewService(fixture.deps()), fixture.resources, generator)
	current := validWorkerCreationDraft()
	current.WorkerSpec.ModelResourceID = 0
	current.WorkerSpec.TypeConfig.Values = map[string]any{}
	current.WorkerSpec.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{}
	current.WorkerSpec.TypeConfig.InteractionMode = specdomain.InteractionModeACP

	result, err := filler.Fill(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		"Configure Cursor",
		101,
		&current,
	)

	require.NoError(t, err)
	assert.Zero(t, result.Draft.WorkerSpec.ModelResourceID)
	assert.Equal(t, "cursor-worker", result.Draft.WorkerSpec.Metadata.Alias)
	require.NotNil(t, generator.resource)
	assert.Equal(t, int64(101), generator.resource.Resource.ID)
}

func TestDraftFillerRequiresCurrentDraftAndPropagatesGeneratorFailure(t *testing.T) {
	_, _, filler := newDraftFillerFixture(`{}`)

	_, err := filler.Fill(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		"Fill the worker",
		101,
		nil,
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)

	_, generator, filler := newDraftFillerFixture(`{}`)
	generator.err = errors.New("provider unavailable")
	current := validWorkerCreationDraft()
	_, err = filler.Fill(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		"Fill the worker",
		101,
		&current,
	)
	assert.ErrorContains(t, err, "provider unavailable")
}

type recordingDraftJSONGenerator struct {
	output       []byte
	err          error
	resource     *resourceservice.ResolvedResource
	systemPrompt string
	userPrompt   string
}

func (generator *recordingDraftJSONGenerator) Generate(
	_ context.Context,
	resource *resourceservice.ResolvedResource,
	systemPrompt, userPrompt string,
) ([]byte, error) {
	generator.resource = resource
	generator.systemPrompt = systemPrompt
	generator.userPrompt = userPrompt
	return generator.output, generator.err
}

func newDraftFillerFixture(
	output string,
) (*workerCreationServiceFixture, *recordingDraftJSONGenerator, *DraftFiller) {
	fixture := newWorkerCreationServiceFixture()
	provider, exists := resourcedomain.Provider("openai")
	if !exists {
		panic("openai provider is missing")
	}
	fixture.resources.resolved.Provider = provider
	fixture.resources.resolved.Connection.BaseURL = provider.DefaultBaseURL
	fixture.resources.resolved.Credentials = map[string]string{"api_key": "must-not-leak"}
	generator := &recordingDraftJSONGenerator{output: []byte(strings.TrimSpace(output))}
	service := NewService(fixture.deps())
	return fixture, generator, NewDraftFiller(service, fixture.resources, generator)
}
