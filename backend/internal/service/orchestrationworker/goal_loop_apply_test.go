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

func TestGoalLoopApplyServiceBuildsPinnedDomainProjection(t *testing.T) {
	state := goalLoopApplyCreateState(t)
	repository := &goalLoopApplyRepositoryStub{
		state: state, goalLoopID: 1001,
	}
	service, err := NewGoalLoopApplyService(
		workerApplyRegistry(t),
		repository,
	)
	require.NoError(t, err)

	applied, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, int64(1001), applied.GoalLoopID)
	assert.Equal(t, int64(901), applied.WorkerSpecSnapshotID)
	assert.Equal(t, int64(1), applied.ResourceRevision)
	assert.Equal(t, "Checkout Recovery", repository.mutation.Projection.Name)
	assert.Equal(t, "Restore checkout reliability", repository.mutation.Projection.Description)
	assert.Equal(t, "Fix checkout", repository.mutation.Projection.Objective)
	assert.Equal(t, []string{"Tests pass", "Evidence recorded"},
		repository.mutation.Projection.AcceptanceCriteria)
	assert.Equal(t, "go test ./...", repository.mutation.Projection.VerificationCommand)
	assert.Equal(t, 100, repository.mutation.Projection.MaxIterations)
	assert.Equal(t, int64(200000), *repository.mutation.Projection.TokenBudget)
	assert.Equal(t, 1440, repository.mutation.Projection.TimeoutMinutes)
	assert.Equal(t, 20, repository.mutation.Projection.NoProgressLimit)
	assert.Equal(t, 20, repository.mutation.Projection.SameErrorLimit)
	assert.Equal(t, "fail", repository.mutation.Projection.EscalationPolicy)
	assert.Equal(t, int64(901), repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestGoalLoopApplyServiceRejectsWrongTargetOrArtifactKind(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*controlservice.LockedApplyState)
	}{
		{
			name: "wrong target kind",
			mutate: func(state *controlservice.LockedApplyState) {
				state.Plan.Target.Kind = resource.KindExpert
			},
		},
		{
			name: "wrong artifact kind",
			mutate: func(state *controlservice.LockedApplyState) {
				state.Plan.ArtifactKind = resource.KindExpert + "Apply"
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			state := goalLoopApplyCreateState(t)
			test.mutate(&state)
			repository := &goalLoopApplyRepositoryStub{state: state}
			service, err := NewGoalLoopApplyService(
				workerApplyRegistry(t),
				repository,
			)
			require.NoError(t, err)

			_, err = service.Apply(
				context.Background(),
				state.Plan.Scope,
				"11111111-1111-4111-8111-111111111111",
			)

			assert.ErrorIs(t, err, control.ErrInvalid)
		})
	}
}

func TestGoalLoopApplyServiceRejectsMissingDependencies(t *testing.T) {
	_, err := NewGoalLoopApplyService(nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
}

type goalLoopApplyRepositoryStub struct {
	state      controlservice.LockedApplyState
	mutation   GoalLoopApplyMutation
	goalLoopID int64
	calls      int
}

func (stub *goalLoopApplyRepositoryStub) RunGoalLoopApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build GoalLoopApplyBuilder,
) (AppliedGoalLoop, error) {
	stub.calls++
	mutation, err := build(stub.state)
	if err != nil {
		return AppliedGoalLoop{}, err
	}
	stub.mutation = mutation
	return AppliedGoalLoop{
		Head: mutation.Head, GoalLoopID: stub.goalLoopID,
		WorkerSpecSnapshotID: mutation.Projection.WorkerSpecSnapshotID,
		ResourceRevision:     mutation.Head.Revision,
	}, nil
}

func goalLoopApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	state := expertApplyCreateState(t)
	workerRef := resource.Reference{
		Kind: resource.KindWorkerTemplate, Name: "review-worker",
	}
	tokenBudget := int64(200000)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindGoalLoop,
		},
		Metadata: resource.Metadata{
			Name: "checkout-recovery", Namespace: "team-alpha",
			DisplayName: "Checkout Recovery", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.GoalLoopResourceSpec{
			WorkerTemplateRef: workerRef,
			Description:       "Restore checkout reliability",
			Objective:         "Fix checkout",
			AcceptanceCriteria: []string{
				"Tests pass",
				"Evidence recorded",
			},
			VerificationCommand: "go test ./...",
			MaxIterations:       100,
			TokenBudget:         &tokenBudget,
			TimeoutMinutes:      1440,
			NoProgressLimit:     20,
			SameErrorLimit:      20,
			EscalationPolicy:    "fail",
		}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = resource.KindGoalLoop + "Apply"
	state.Plan.ArtifactJSON = canonicalApplyJSON(t, DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 901,
	})
	state.Plan.ArtifactDigest = digestApplyJSON(t, state.Plan.ArtifactJSON)
	state.Plan.ResolvedReferences = []control.ResolvedReference{
		resolvedApplyReference(state.Plan.Scope, workerRef, 2, "a"),
	}
	state.ResultIdentity.ResourceTarget = state.Plan.Target
	return state
}
