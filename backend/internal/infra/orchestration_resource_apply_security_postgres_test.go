package infra

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestrationApplyTransactionUsesDatabaseTimeForExpiry(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationTestCreatePlan(t)
	plan.CreatedAt = time.Now().UTC().Add(-2 * time.Hour)
	plan.ExpiresAt = plan.CreatedAt.Add(time.Hour)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			t.Fatal("expired plan must not invoke builder")
			return orchestrationservice.ApplyMutation{}, nil
		},
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrExpired)
}

func TestOrchestrationApplyTransactionRejectsUnplannedManifest(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(state orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			mutation := orchestrationCreateMutation(t, state)
			var manifest map[string]any
			require.NoError(t, json.Unmarshal(
				mutation.Revision.CanonicalManifest,
				&manifest,
			))
			manifest["spec"] = map[string]any{"mode": "unplanned"}
			canonical, canonicalErr := orchestrationcontrol.CanonicalJSONObject(manifest)
			require.NoError(t, canonicalErr)
			spec, specErr := orchestrationcontrol.CanonicalJSONObject(manifest["spec"])
			require.NoError(t, specErr)
			digest, digestErr := orchestrationcontrol.DigestCanonicalJSON(canonical)
			require.NoError(t, digestErr)
			mutation.Revision.CanonicalManifest = canonical
			mutation.Revision.CanonicalSpec = spec
			mutation.Revision.Digest = digest
			return mutation, nil
		},
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrInvalid)

	stored, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
}

func TestOrchestrationApplyTransactionRejectsUnplannedReferences(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(state orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			mutation := orchestrationCreateMutation(t, state)
			mutation.Revision.ResolvedReferences = []orchestrationcontrol.ResolvedReference{}
			return mutation, nil
		},
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrInvalid)
}
