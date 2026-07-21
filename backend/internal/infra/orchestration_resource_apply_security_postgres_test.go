package infra

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	orchestrationservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func TestOrchestrationApplyTransactionRechecksRevokedMembership(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	require.NoError(t, db.Where(
		"organization_id = ? AND user_id = ?",
		plan.Scope.OrganizationID,
		plan.Scope.ActorID,
	).Delete(&organization.Member{}).Error)

	_, err := repo.RunApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(orchestrationservice.LockedApplyState) (
			orchestrationservice.ApplyMutation,
			error,
		) {
			t.Fatal("revoked member must not invoke builder")
			return orchestrationservice.ApplyMutation{}, nil
		},
	)

	assert.ErrorIs(t, err, orchestrationservice.ErrForbidden)
	stored, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
	var resourceCount, revisionCount int64
	require.NoError(t, db.Table("orchestration_resources").Count(&resourceCount).Error)
	require.NoError(t, db.Table("orchestration_resource_revisions").Count(&revisionCount).Error)
	assert.Zero(t, resourceCount)
	assert.Zero(t, revisionCount)
}

func TestOrchestrationApplyTransactionWaitsForConcurrentRevocation(
	t *testing.T,
) {
	db, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	revoke := db.Begin()
	require.NoError(t, revoke.Error)
	defer revoke.Rollback()
	require.NoError(t, revoke.Where(
		"organization_id = ? AND user_id = ?",
		plan.Scope.OrganizationID,
		plan.Scope.ActorID,
	).Delete(&organization.Member{}).Error)
	var blockerPID int
	require.NoError(t, revoke.Raw(
		"SELECT pg_backend_pid()",
	).Scan(&blockerPID).Error)
	applyCtx, cancelApply := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelApply()

	result := make(chan error, 1)
	go func() {
		_, err := repo.RunApplyTransaction(
			applyCtx,
			plan.Scope,
			plan.ID,
			func(state orchestrationservice.LockedApplyState) (
				orchestrationservice.ApplyMutation,
				error,
			) {
				return orchestrationservice.ApplyMutation{}, assert.AnError
			},
		)
		result <- err
	}()

	requireApplyBlockedOnMemberLock(t, db, blockerPID)
	require.NoError(t, revoke.Commit().Error)
	assert.ErrorIs(
		t,
		requireApplyResult(t, result),
		orchestrationservice.ErrForbidden,
	)
	stored, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
	var resourceCount, revisionCount int64
	require.NoError(t, db.Table("orchestration_resources").Count(&resourceCount).Error)
	require.NoError(t, db.Table("orchestration_resource_revisions").Count(&revisionCount).Error)
	assert.Zero(t, resourceCount)
	assert.Zero(t, revisionCount)
}

func TestOrchestrationApplyReplayRechecksRevokedMembership(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	builder := func(state orchestrationservice.LockedApplyState) (
		orchestrationservice.ApplyMutation,
		error,
	) {
		return orchestrationCreateMutation(t, state), nil
	}
	_, err := repo.runIdempotentApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		plan.Target.Kind,
		plan.ArtifactKind,
		builder,
	)
	require.NoError(t, err)
	require.NoError(t, db.Where(
		"organization_id = ? AND user_id = ?",
		plan.Scope.OrganizationID,
		plan.Scope.ActorID,
	).Delete(&organization.Member{}).Error)

	_, err = repo.runIdempotentApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		plan.Target.Kind,
		plan.ArtifactKind,
		builder,
	)

	assert.ErrorIs(t, err, orchestrationservice.ErrForbidden)
}

func TestOrchestrationUpdateApplyWaitsForConcurrentRoleDowngrade(
	t *testing.T,
) {
	db, repo := orchestrationPostgresRepository(t)
	require.NoError(t, db.Create(&organization.Member{
		OrganizationID: 42,
		UserID:         8,
		Role:           organization.RoleOwner,
	}).Error)
	head := orchestrationApplyInitialResourceForActor(t, repo, 8)
	plan := orchestrationUpdatePlan(t, head, "strict")
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	require.NoError(t, db.Model(&organization.Member{}).
		Where("organization_id = ? AND user_id = ?", 42, 7).
		Update("role", organization.RoleAdmin).Error)
	downgrade := db.Begin()
	require.NoError(t, downgrade.Error)
	defer downgrade.Rollback()
	require.NoError(t, downgrade.Model(&organization.Member{}).
		Where("organization_id = ? AND user_id = ?", 42, 7).
		Update("role", organization.RoleMember).Error)
	var blockerPID int
	require.NoError(t, downgrade.Raw(
		"SELECT pg_backend_pid()",
	).Scan(&blockerPID).Error)
	applyCtx, cancelApply := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelApply()

	result := make(chan error, 1)
	go func() {
		_, err := repo.RunApplyTransaction(
			applyCtx,
			plan.Scope,
			plan.ID,
			func(orchestrationservice.LockedApplyState) (
				orchestrationservice.ApplyMutation,
				error,
			) {
				return orchestrationservice.ApplyMutation{}, assert.AnError
			},
		)
		result <- err
	}()

	requireApplyBlockedOnMemberLock(t, db, blockerPID)
	require.NoError(t, downgrade.Commit().Error)
	assert.ErrorIs(
		t,
		requireApplyResult(t, result),
		orchestrationservice.ErrForbidden,
	)
	stored, loadErr := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
}

func requireApplyBlockedOnMemberLock(
	t *testing.T,
	db *gorm.DB,
	blockerPID int,
) {
	t.Helper()
	require.Eventually(t, func() bool {
		var count int64
		err := db.Raw(`
SELECT count(*)
FROM pg_stat_activity
WHERE ? = ANY(pg_blocking_pids(pid))
  AND query ILIKE '%organization_members%'`,
			blockerPID,
		).Scan(&count).Error
		return err == nil && count > 0
	}, 3*time.Second, 20*time.Millisecond)
}

func requireApplyResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("apply transaction did not finish")
		return nil
	}
}
