package infra

import (
	"context"
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestrationResourceRepositoryScopesResourceReads(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	scope := orchestrationTestScope()
	head := orchestrationTestHead()
	insertOrchestrationHead(t, db, head)
	foreign := head
	foreign.ID = 202
	foreign.OrganizationID = 99
	foreign.Identity.UID = "44444444-4444-4444-8444-444444444444"
	foreign.Identity.Name = "worker-two"
	insertOrchestrationHead(t, db, foreign)

	loaded, err := repo.GetResource(context.Background(), scope, head.Identity.ResourceTarget)
	require.NoError(t, err)
	assert.Equal(t, head, loaded)

	foreignScope := scope
	foreignScope.OrganizationID = 99
	foreignScope.OrganizationSlug = "team-beta"
	foreignTarget := head.Identity.ResourceTarget
	foreignTarget.Namespace = "team-beta"
	_, err = repo.GetResource(context.Background(), foreignScope, foreignTarget)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrNotFound)

	page, err := repo.ListResources(context.Background(), scope, orchestrationservice.ResourceListFilter{
		Kind:   "WorkerTemplate",
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, int64(1), page.Total)
	assert.Equal(t, head.ID, page.Items[0].ID)
}

func TestOrchestrationResourceRepositoryScopesRevisionReads(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	scope := orchestrationTestScope()
	head := orchestrationTestHead()
	revision := orchestrationTestRevision(t, head)
	insertOrchestrationHead(t, db, head)
	insertOrchestrationRevision(t, db, revision)

	loaded, err := repo.GetRevision(context.Background(), scope, head.ID, revision.Revision)
	require.NoError(t, err)
	assert.Equal(t, revision, loaded)

	_, err = repo.GetRevision(context.Background(), scope, head.ID+1, revision.Revision)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrNotFound)
	revisions, err := repo.ListRevisions(context.Background(), scope, head.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
	assert.Equal(t, revision.Digest, revisions[0].Digest)
}

func TestOrchestrationResourceRepositoryBindsPlansToTenantAndActor(t *testing.T) {
	_, repo := orchestrationRepositoryForTest(t)
	plan := orchestrationTestCreatePlan(t)

	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	assert.ErrorIs(t, repo.CreatePlan(context.Background(), plan), orchestrationcontrol.ErrConflict)

	loaded, err := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, plan, loaded)

	otherActor := plan.Scope
	otherActor.ActorID++
	_, err = repo.GetPlan(context.Background(), otherActor, plan.ID)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrNotFound)
	otherTenant := plan.Scope
	otherTenant.OrganizationID++
	otherTenant.OrganizationSlug = "team-beta"
	_, err = repo.GetPlan(context.Background(), otherTenant, plan.ID)
	assert.ErrorIs(t, err, orchestrationcontrol.ErrNotFound)
}

func TestOrchestrationResourceRepositoryRejectsCorruptStoredJSON(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	head := orchestrationTestHead()
	insertOrchestrationHead(t, db, head)
	require.NoError(t, db.Exec(
		`UPDATE orchestration_resources SET status = CAST('[]' AS BLOB) WHERE id = ?`,
		head.ID,
	).Error)

	_, err := repo.GetResource(
		context.Background(),
		orchestrationTestScope(),
		head.Identity.ResourceTarget,
	)
	require.Error(t, err)
	assert.True(t, errors.Is(err, orchestrationcontrol.ErrCorrupt))
	assert.False(t, errors.Is(err, orchestrationcontrol.ErrNotFound))
}

func TestOrchestrationResourceRepositoryRejectsCorruptPlanState(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	plan := orchestrationTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	require.NoError(t, db.Exec(`
UPDATE orchestration_resource_plans
SET result_json = CAST('{}' AS BLOB)
WHERE id = ?`, plan.ID).Error)

	_, err := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.ErrorIs(t, err, orchestrationcontrol.ErrCorrupt)
}

func TestOrchestrationResourceRepositoryRejectsInvalidPlanID(t *testing.T) {
	_, repo := orchestrationRepositoryForTest(t)
	_, err := repo.GetPlan(
		context.Background(),
		orchestrationTestScope(),
		"not-a-plan-id",
	)
	require.ErrorIs(t, err, orchestrationcontrol.ErrInvalid)
}
