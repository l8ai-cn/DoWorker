package infra

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestrationApplyTransactionCreatesAndConsumesAtomically(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))

	head, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(state orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			require.Nil(t, state.Head)
			require.Nil(t, state.CurrentRevision)
			require.Positive(t, state.ResultResourceID)
			require.NotEmpty(t, state.ResultIdentity.UID)
			return orchestrationCreateMutation(t, state), nil
		},
	)
	require.NoError(t, err)
	require.Positive(t, head.ID)

	stored, err := repo.GetResource(
		context.Background(),
		plan.Scope,
		plan.Target,
	)
	require.NoError(t, err)
	assert.Equal(t, head, stored)
	revision, err := repo.GetRevision(
		context.Background(),
		plan.Scope,
		head.ID,
		1,
	)
	require.NoError(t, err)
	assert.Equal(t, head.Identity, revision.Identity)
	storedPlan, err := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, orchestrationcontrol.PlanStatusApplied, storedPlan.Status)
	assert.Equal(t, head.ID, storedPlan.ResultResourceID)

	var resourceCount, revisionCount int64
	require.NoError(t, db.Table("orchestration_resources").Count(&resourceCount).Error)
	require.NoError(t, db.Table("orchestration_resource_revisions").Count(&revisionCount).Error)
	assert.Equal(t, int64(1), resourceCount)
	assert.Equal(t, int64(1), revisionCount)
}

func TestOrchestrationApplyTransactionRollsBackBuilderFailure(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	builderErr := errors.New("artifact persistence failed")

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			return orchestrationservice.ApplyMutation{}, builderErr
		},
	)
	assert.ErrorIs(t, err, builderErr)

	storedPlan, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, storedPlan.Status)
	var resourceCount int64
	require.NoError(t, db.Table("orchestration_resources").Count(&resourceCount).Error)
	assert.Zero(t, resourceCount)
}

func TestOrchestrationApplyTransactionRejectsReplay(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	builder := func(state orchestrationservice.LockedApplyState) (
		orchestrationservice.ApplyMutation,
		error,
	) {
		return orchestrationCreateMutation(t, state), nil
	}
	_, err := repo.RunApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)
	_, err = repo.RunApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrConsumed)
}

func TestOrchestrationApplyTransactionSerializesConcurrentConsumption(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			_, err := repo.RunApplyTransaction(
				context.Background(),
				plan.Scope,
				plan.ID,
				func(state orchestrationservice.LockedApplyState) (
					orchestrationservice.ApplyMutation,
					error,
				) {
					return orchestrationCreateMutation(t, state), nil
				},
			)
			results <- err
		}()
	}
	ready.Wait()
	close(start)
	first, second := <-results, <-results
	assert.True(t,
		(first == nil && errors.Is(second, orchestrationcontrol.ErrConsumed)) ||
			(second == nil && errors.Is(first, orchestrationcontrol.ErrConsumed)),
		"results: %v, %v", first, second,
	)
}

func TestOrchestrationApplyTransactionRejectsExpiredPlan(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
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

func TestOrchestrationApplyTransactionUpdatesExactBase(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	head := orchestrationApplyInitialResource(t, repo)
	plan := orchestrationUpdatePlan(t, head, "strict")
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	updated, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(state orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			require.NotNil(t, state.Head)
			require.NotNil(t, state.CurrentRevision)
			require.Equal(t, head.ResourceVersion, state.Head.ResourceVersion)
			return orchestrationUpdateMutation(t, state), nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), updated.Revision)
	assert.Equal(t, int64(2), updated.Generation)
	assert.Equal(t, int64(2), updated.ResourceVersion)

	revision, err := repo.GetRevision(
		context.Background(),
		plan.Scope,
		head.ID,
		2,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), revision.Generation)
}

func TestOrchestrationApplyTransactionRejectsStaleBaseBeforeBuilder(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	head := orchestrationApplyInitialResource(t, repo)
	plan := orchestrationUpdatePlan(t, head, "strict")
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	require.NoError(t, db.Exec(`
UPDATE orchestration_resources
SET status = '{"ready":true}', resource_version = resource_version + 1,
	updated_by_id = 7, updated_at = updated_at + interval '1 second'
WHERE id = ?`, head.ID).Error)

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			t.Fatal("stale plan must not invoke builder")
			return orchestrationservice.ApplyMutation{}, nil
		},
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrStale)

	stored, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
}

func TestOrchestrationApplyTransactionRejectsCreateTargetConflict(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	first := orchestrationApplyTestCreatePlan(t)
	second := first
	second.ID = "55555555-5555-4555-8555-555555555555"
	require.NoError(t, repo.CreatePlan(context.Background(), first))
	require.NoError(t, repo.CreatePlan(context.Background(), second))
	_, err := repo.RunApplyTransaction(
		context.Background(),
		first.Scope,
		first.ID,
		func(state orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			return orchestrationCreateMutation(t, state), nil
		},
	)
	require.NoError(t, err)
	_, err = repo.RunApplyTransaction(
		context.Background(),
		second.Scope,
		second.ID,
		func(orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			t.Fatal("conflicting create must not invoke builder")
			return orchestrationservice.ApplyMutation{}, nil
		},
	)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrConflict)
}

func TestOrchestrationApplyTransactionSerializesCompetingUpdates(t *testing.T) {
	_, repo := orchestrationPostgresRepository(t)
	head := orchestrationApplyInitialResource(t, repo)
	first := orchestrationUpdatePlan(t, head, "strict")
	second := orchestrationUpdatePlan(t, head, "review")
	second.ID = "66666666-6666-4666-8666-666666666666"
	require.NoError(t, repo.CreatePlan(context.Background(), first))
	require.NoError(t, repo.CreatePlan(context.Background(), second))
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, plan := range []orchestrationcontrol.Plan{first, second} {
		plan := plan
		go func() {
			<-start
			_, err := repo.RunApplyTransaction(
				context.Background(),
				plan.Scope,
				plan.ID,
				func(state orchestrationservice.LockedApplyState) (
					orchestrationservice.ApplyMutation,
					error,
				) {
					return orchestrationUpdateMutation(t, state), nil
				},
			)
			results <- err
		}()
	}
	close(start)
	firstErr, secondErr := <-results, <-results
	assert.True(t,
		(firstErr == nil && errors.Is(secondErr, orchestrationcontrol.ErrStale)) ||
			(secondErr == nil && errors.Is(firstErr, orchestrationcontrol.ErrStale)),
		"results: %v, %v", firstErr, secondErr,
	)
}
