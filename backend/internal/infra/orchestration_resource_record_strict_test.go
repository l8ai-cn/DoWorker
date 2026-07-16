package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/stretchr/testify/require"
)

func TestOrchestrationResourceRepositoryRejectsUnknownPlanJSONFields(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	plan := orchestrationTestCreatePlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	require.NoError(t, db.Exec(`
UPDATE orchestration_resource_plans
SET resolved_refs = CAST('[{"apiVersion":"agentsmesh.io/v1alpha1",
	"kind":"ModelBinding","namespace":"team-alpha","name":"coding-primary",
	"uid":"33333333-3333-4333-8333-333333333333","revision":1,
	"digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	"unexpected":true}]' AS BLOB)
WHERE id = ?`, plan.ID).Error)

	_, err := repo.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.ErrorIs(t, err, orchestrationcontrol.ErrCorrupt)
}

func TestOrchestrationResourceRepositoryRejectsUnknownRevisionJSONFields(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	head := orchestrationTestHead()
	revision := orchestrationTestRevision(t, head)
	insertOrchestrationHead(t, db, head)
	insertOrchestrationRevision(t, db, revision)
	require.NoError(t, db.Exec(`
UPDATE orchestration_resource_revisions
SET resolved_refs = CAST('[{"apiVersion":"agentsmesh.io/v1alpha1",
	"kind":"ModelBinding","namespace":"team-alpha","name":"coding-primary",
	"uid":"33333333-3333-4333-8333-333333333333","revision":1,
	"digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	"unexpected":true}]' AS BLOB)
WHERE resource_id = ?`, head.ID).Error)

	_, err := repo.GetRevision(
		context.Background(),
		orchestrationTestScope(),
		head.ID,
		revision.Revision,
	)
	require.ErrorIs(t, err, orchestrationcontrol.ErrCorrupt)
}
