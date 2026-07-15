package infra

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindingApplyPersistsRevisionWithoutWorkerSpecSnapshot(t *testing.T) {
	_, repository := orchestrationPostgresRepository(t)
	plan := orchestrationBindingApplyPlan(t)
	require.NoError(t, repository.CreatePlan(context.Background(), plan))

	head, err := repository.RunBindingApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			return orchestrationCreateMutation(t, state), nil
		},
	)

	require.NoError(t, err)
	revision, err := repository.GetRevision(
		context.Background(),
		plan.Scope,
		head.ID,
		1,
	)
	require.NoError(t, err)
	assert.Zero(t, revision.WorkerSpecSnapshotID)
}

func TestBindingApplyRejectsWorkerTemplateBeforeBuilder(t *testing.T) {
	_, repository := orchestrationPostgresRepository(t)
	plan := orchestrationWorkerTemplateApplyPlan(t)
	require.NoError(t, repository.CreatePlan(context.Background(), plan))

	_, err := repository.RunBindingApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(
			controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			t.Fatal("WorkerTemplate must not invoke a binding builder")
			return controlservice.ApplyMutation{}, nil
		},
	)

	assert.ErrorIs(t, err, control.ErrInvalid)
}

func orchestrationBindingApplyPlan(t *testing.T) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	plan.Target.Kind = resource.KindModelBinding
	var manifest resource.Manifest
	require.NoError(t, json.Unmarshal(plan.CanonicalManifest, &manifest))
	manifest.Kind = resource.KindModelBinding
	plan.CanonicalManifest = mustCanonicalObject(t, manifest)
	plan.ArtifactKind = resource.KindModelBinding + "Spec"
	plan.ArtifactJSON = mustCanonicalObject(t, map[string]any{
		"provider": "openai",
		"model":    "gpt-5",
	})
	var err error
	plan.DraftHash, err = control.DigestCanonicalJSON(plan.CanonicalManifest)
	require.NoError(t, err)
	plan.ArtifactDigest, err = control.DigestCanonicalJSON(plan.ArtifactJSON)
	require.NoError(t, err)
	plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	return plan
}

var _ workerplanner.BindingApplyRepository = (*orchestrationResourceRepo)(nil)
