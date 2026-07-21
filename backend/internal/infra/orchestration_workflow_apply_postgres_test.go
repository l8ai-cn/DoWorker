package infra

import (
	"context"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowApplyTransactionCreatesPinnedProjectionAndReplays(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	applyOrchestrationDomainLinkFixtures(t, db)
	plan := orchestrationWorkflowApplyPlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	nextRunAt := time.Date(2026, 7, 16, 2, 0, 0, 0, time.UTC)
	builder := func(state controlservice.LockedApplyState) (
		workerplanner.WorkflowApplyMutation,
		error,
	) {
		mutation := orchestrationCreateMutation(t, state)
		mutation.Revision.WorkerSpecSnapshotID = 100
		return workerplanner.WorkflowApplyMutation{
			ApplyMutation: mutation,
			Projection: workerplanner.WorkflowApplyProjection{
				Name: "Nightly Review", Prompt: "Review authorization",
				ExecutionMode: "direct", CronExpression: "0 2 * * *",
				SandboxStrategy: "fresh", ConcurrencyPolicy: "skip",
				MaxConcurrentRuns: 1, MaxRetainedRuns: 30,
				TimeoutMinutes: 60, IdleTimeoutSeconds: 30,
				WorkerSpecSnapshotID: 100, NextRunAt: &nextRunAt,
			},
		}, nil
	}

	first, err := repo.RunWorkflowApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)
	replayed, err := repo.RunWorkflowApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)

	assert.Equal(t, first, replayed)
	assert.Equal(t, int64(100), first.WorkerSpecSnapshotID)
	var count int64
	require.NoError(t, db.Table("workflows").Count(&count).Error)
	assert.Equal(t, int64(1), count)
	var resourceID, revision, snapshotID int64
	var prompt string
	require.NoError(t, db.Table("workflows").
		Select("orchestration_resource_id, orchestration_resource_revision, worker_spec_snapshot_id, prompt_template").
		Row().Scan(&resourceID, &revision, &snapshotID, &prompt))
	assert.Equal(t, first.Head.ID, resourceID)
	assert.Equal(t, first.Head.Revision, revision)
	assert.Equal(t, first.WorkerSpecSnapshotID, snapshotID)
	assert.Equal(t, "Review authorization", prompt)
}

func orchestrationWorkflowApplyPlan(t *testing.T) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkflow,
		},
		Metadata: resource.Metadata{
			Name: "nightly-review", Namespace: "team-alpha",
			DisplayName: "Nightly Review", Labels: map[string]string{},
		},
		Spec: expertCanonicalJSON(t, resource.WorkflowResourceSpec{
			WorkerTemplateRef: resource.Reference{
				Kind: resource.KindWorkerTemplate, Name: "review-worker",
			},
			PromptRef: resource.Reference{
				Kind: resource.KindPrompt, Name: "review-prompt",
			},
			Inputs:        map[string]string{"scope": "authorization"},
			ExecutionMode: "direct", CronExpression: "0 2 * * *",
			SandboxStrategy: "fresh", ConcurrencyPolicy: "skip",
			MaxConcurrentRuns: 1, MaxRetainedRuns: 30,
			TimeoutMinutes: 60, IdleTimeoutSeconds: 30,
		}),
	}
	plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	plan.CanonicalManifest = expertCanonicalJSON(t, manifest)
	plan.DraftHash = expertDigestJSON(t, plan.CanonicalManifest)
	plan.ArtifactKind = resource.KindWorkflow + "Apply"
	plan.ArtifactJSON = expertCanonicalJSON(t, workerplanner.DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 100,
	})
	plan.ArtifactDigest = expertDigestJSON(t, plan.ArtifactJSON)
	plan.ResolvedReferences = []control.ResolvedReference{}
	plan.PlanHash = expertPlanHash(t, plan)
	return plan
}
