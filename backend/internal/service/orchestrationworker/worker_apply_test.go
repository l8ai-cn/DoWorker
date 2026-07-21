package orchestrationworker

import (
	"context"
	"testing"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildWorkerApplyMutationPinsRenderedInvocation(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	resolver := &definitionResolverStub{
		prompt: resource.PromptSpec{
			Content: "Review {{scope}} {{tone}}",
			Variables: map[string]resource.PromptVariableSpec{
				"scope": {Required: true},
				"tone":  {Default: stringPointerForTest("carefully")},
			},
		},
	}

	mutation, err := buildWorkerApplyMutation(
		context.Background(),
		workerApplyRegistry(t),
		resolver,
		state,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(901), mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, int64(901), mutation.Launch.WorkerSpecSnapshotID)
	require.NotNil(t, mutation.Launch.Prompt)
	assert.Equal(t, "Review authorization carefully", *mutation.Launch.Prompt)
	assert.Equal(t, "reviewer-42", mutation.Launch.Alias)
}

func TestBuildWorkerApplyMutationRejectsUpdate(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	state.Plan.Operation = control.PlanOperationUpdate

	_, err := buildWorkerApplyMutation(
		context.Background(),
		workerApplyRegistry(t),
		&definitionResolverStub{},
		state,
	)

	assert.ErrorIs(t, err, control.ErrInvalid)
}

func workerInvocationApplyCreateState(t *testing.T) controlservice.LockedApplyState {
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
			Kind:       resource.KindWorker,
		},
		Metadata: resource.Metadata{
			Name: "reviewer-42", Namespace: "team-alpha",
			DisplayName: "Reviewer 42", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.WorkerInvocationSpec{
			WorkerTemplateRef: workerRef, PromptRef: &promptRef,
			Inputs: map[string]string{"scope": "authorization"},
			Alias:  "reviewer-42",
		}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = resource.KindWorker + "Apply"
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
