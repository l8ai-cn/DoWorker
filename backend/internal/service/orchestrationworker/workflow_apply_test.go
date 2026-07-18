package orchestrationworker

import (
	"context"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowApplyServiceBuildsPinnedDomainProjection(t *testing.T) {
	state := workflowApplyCreateState(t)
	repository := &workflowApplyRepositoryStub{
		state: state, workflowID: 801,
	}
	resolver := &definitionResolverStub{
		prompt: resource.PromptSpec{
			Content: "Review {{scope}} {{tone}}",
			Variables: map[string]resource.PromptVariableSpec{
				"scope": {Required: true},
				"tone":  {Default: stringPointerForTest("carefully")},
			},
		},
	}
	service, err := NewWorkflowApplyService(
		workerApplyRegistry(t),
		repository,
		resolver,
	)
	require.NoError(t, err)

	applied, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, int64(801), applied.WorkflowID)
	assert.Equal(t, int64(901), applied.WorkerSpecSnapshotID)
	assert.Equal(t, int64(1), applied.ResourceRevision)
	assert.Equal(t, "Review authorization carefully", repository.mutation.Projection.Prompt)
	assert.Equal(t, "enabled", repository.mutation.Projection.Status)
	assert.Equal(t, "direct", repository.mutation.Projection.ExecutionMode)
	assert.Equal(t, "0 2 * * *", repository.mutation.Projection.CronExpression)
	assert.Equal(t, int64(901), repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestWorkflowApplyServiceBuildsDisabledProjectionAtomically(t *testing.T) {
	state := workflowApplyCreateState(t)
	repository := &workflowApplyRepositoryStub{state: state, workflowID: 802}
	service, err := NewWorkflowApplyService(
		workerApplyRegistry(t),
		repository,
		&definitionResolverStub{prompt: resource.PromptSpec{
			Content: "Review {{scope}}",
			Variables: map[string]resource.PromptVariableSpec{
				"scope": {Required: true},
			},
		}},
	)
	require.NoError(t, err)

	_, err = service.ApplyWithStatus(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
		"disabled",
	)

	require.NoError(t, err)
	assert.Equal(t, "disabled", repository.mutation.Projection.Status)
	assert.Nil(t, repository.mutation.Projection.NextRunAt)
}

func TestWorkflowApplyServiceRejectsMissingDependencies(t *testing.T) {
	_, err := NewWorkflowApplyService(nil, nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
}

type workflowApplyRepositoryStub struct {
	state      controlservice.LockedApplyState
	mutation   WorkflowApplyMutation
	workflowID int64
	calls      int
}

func (stub *workflowApplyRepositoryStub) RunWorkflowApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build WorkflowApplyBuilder,
) (AppliedWorkflow, error) {
	stub.calls++
	mutation, err := build(stub.state)
	if err != nil {
		return AppliedWorkflow{}, err
	}
	stub.mutation = mutation
	return AppliedWorkflow{
		Head: mutation.Head, WorkflowID: stub.workflowID,
		WorkerSpecSnapshotID: mutation.Projection.WorkerSpecSnapshotID,
		ResourceRevision:     mutation.Head.Revision,
	}, nil
}

func workflowApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	state := expertApplyCreateState(t)
	workerRef := resource.Reference{
		Kind: resource.KindWorkerTemplate, Name: "review-worker",
	}
	promptRef := resource.Reference{
		Kind: resource.KindPrompt, Name: "review-system",
	}
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkflow,
		},
		Metadata: resource.Metadata{
			Name: "nightly-review", Namespace: "team-alpha",
			DisplayName: "Nightly Review", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.WorkflowResourceSpec{
			WorkerTemplateRef: workerRef, PromptRef: promptRef,
			Inputs:        map[string]string{"scope": "authorization"},
			ExecutionMode: "direct", CronExpression: "0 2 * * *",
			SandboxStrategy: "fresh", SessionPersistence: false,
			ConcurrencyPolicy: "skip", MaxConcurrentRuns: 1,
			MaxRetainedRuns: 30, TimeoutMinutes: 60,
			IdleTimeoutSeconds: 30,
		}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = resource.KindWorkflow + "Apply"
	state.Plan.ArtifactJSON = canonicalApplyJSON(t, DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 901,
	})
	state.Plan.ArtifactDigest = digestApplyJSON(t, state.Plan.ArtifactJSON)
	state.Plan.ResolvedReferences = []control.ResolvedReference{
		resolvedApplyReference(state.Plan.Scope, workerRef, 2, "a"),
		resolvedApplyReference(state.Plan.Scope, promptRef, 3, "b"),
	}
	state.ResultIdentity.ResourceTarget = state.Plan.Target
	return state
}
