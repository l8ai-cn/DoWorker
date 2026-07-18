package infra

import (
	"context"
	"testing"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/migrations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWorkerApplyTransactionCreatesOneShotLaunchAndReplays(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	applyOrchestrationDomainLinkFixtures(t, db)
	applyWorkerLaunchFixtures(t, db)
	plan := orchestrationWorkerApplyPlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	prompt := "Review authorization carefully"
	builder := func(state controlservice.LockedApplyState) (
		workerplanner.WorkerApplyMutation,
		error,
	) {
		mutation := orchestrationCreateMutation(t, state)
		mutation.Revision.WorkerSpecSnapshotID = 100
		return workerplanner.WorkerApplyMutation{
			ApplyMutation: mutation,
			Launch: workerplanner.WorkerLaunchProjection{
				WorkerSpecSnapshotID: 100,
				Prompt:               &prompt,
				Alias:                "reviewer-42",
			},
		}, nil
	}

	first, err := repo.RunWorkerApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)
	replayed, err := repo.RunWorkerApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)

	assert.Equal(t, first, replayed)
	assert.Positive(t, first.LaunchID)
	assert.Equal(t, int64(100), first.WorkerSpecSnapshotID)
	assert.Empty(t, first.PodKey)
	var count int64
	require.NoError(t, db.Table("orchestration_worker_launches").Count(&count).Error)
	assert.Equal(t, int64(1), count)
	var resourceID, revision, snapshotID int64
	var state, storedPrompt, alias string
	require.NoError(t, db.Table("orchestration_worker_launches").
		Select("resource_id, resource_revision, worker_spec_snapshot_id, state, prompt, alias").
		Row().Scan(
		&resourceID,
		&revision,
		&snapshotID,
		&state,
		&storedPrompt,
		&alias,
	))
	assert.Equal(t, first.Head.ID, resourceID)
	assert.Equal(t, first.Head.Revision, revision)
	assert.Equal(t, first.WorkerSpecSnapshotID, snapshotID)
	assert.Equal(t, "pending", state)
	assert.Equal(t, prompt, storedPrompt)
	assert.Equal(t, "reviewer-42", alias)
}

func TestWorkerLaunchClaimCompletesPodAndOutboxAtomically(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	applyOrchestrationDomainLinkFixtures(t, db)
	applyWorkerLaunchFixtures(t, db)
	plan := orchestrationWorkerApplyPlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	applied, err := repo.RunWorkerApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		workerLaunchMutationBuilder(t),
	)
	require.NoError(t, err)
	claim, err := repo.ClaimWorkerLaunch(
		context.Background(),
		plan.Scope,
		applied.LaunchID,
		time.Minute,
		uuid.NewString(),
	)
	require.NoError(t, err)
	assert.Equal(t, applied.LaunchID, claim.LaunchID)
	assert.Equal(t, int64(1), claim.ResourceRevision)
	assert.Equal(t, int64(100), claim.WorkerSpecSnapshotID)
	require.NoError(t, db.Exec(
		"INSERT INTO runners (id, organization_id) VALUES (?, ?)",
		11,
		42,
	).Error)
	require.NoError(t, db.Exec(`
INSERT INTO pods (
	organization_id, pod_key, runner_id, created_by_id,
	worker_spec_snapshot_id, orchestration_worker_launch_id, status
) VALUES (?, ?, ?, ?, ?, ?, 'queued')`,
		42,
		"7-standalone-12345678",
		11,
		7,
		100,
		applied.LaunchID,
	).Error)
	result, err := repo.CompleteWorkerLaunch(
		context.Background(),
		plan.Scope,
		claim,
		workerplanner.WorkerPodLaunch{
			PodID: 1, PodKey: "7-standalone-12345678", RunnerID: 11,
			CommandPayload: encryptedWorkerLaunchPayload(t, "7-standalone-12345678"),
		},
		time.Hour,
	)
	require.NoError(t, err)

	assert.Equal(t, "7-standalone-12345678", result.PodKey)
	assert.Equal(t, int64(11), result.RunnerID)
	var state, commandID string
	var pendingCount int64
	require.NoError(t, db.Table("orchestration_worker_launches").
		Select("state").Where("id = ?", applied.LaunchID).
		Row().Scan(&state))
	assert.Equal(t, "dispatched", state)
	require.NoError(t, db.Table("pending_runner_commands").
		Count(&pendingCount).Error)
	assert.Equal(t, int64(1), pendingCount)
	require.NoError(t, db.Table("pending_runner_commands").
		Select("command_id").Row().Scan(&commandID))
	assert.Equal(t, "worker-1", commandID)
}

func workerLaunchMutationBuilder(
	t *testing.T,
) workerplanner.WorkerApplyBuilder {
	t.Helper()
	return func(state controlservice.LockedApplyState) (
		workerplanner.WorkerApplyMutation,
		error,
	) {
		mutation := orchestrationCreateMutation(t, state)
		mutation.Revision.WorkerSpecSnapshotID = 100
		return workerplanner.WorkerApplyMutation{
			ApplyMutation: mutation,
			Launch: workerplanner.WorkerLaunchProjection{
				WorkerSpecSnapshotID: 100,
				Alias:                "reviewer-42",
			},
		}, nil
	}
}

func applyWorkerLaunchFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
CREATE TABLE runners (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL
);
CREATE TABLE pods (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	pod_key VARCHAR(100) NOT NULL UNIQUE,
	runner_id BIGINT NOT NULL REFERENCES runners(id),
	created_by_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT,
	status VARCHAR(50) NOT NULL DEFAULT 'initializing'
);
CREATE TABLE pending_runner_commands (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	runner_id BIGINT NOT NULL REFERENCES runners(id),
	pod_key VARCHAR(100) NOT NULL,
	command_type VARCHAR(20) NOT NULL,
	command_id VARCHAR(64) NOT NULL UNIQUE,
	payload BYTEA NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`).Error)
	up, err := migrations.FS.ReadFile(
		"000216_orchestration_worker_launches.up.sql",
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
}

func orchestrationWorkerApplyPlan(t *testing.T) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorker,
		},
		Metadata: resource.Metadata{
			Name: "reviewer-42", Namespace: "team-alpha",
			DisplayName: "Reviewer 42", Labels: map[string]string{},
		},
		Spec: expertCanonicalJSON(t, resource.WorkerInvocationSpec{
			WorkerTemplateRef: resource.Reference{
				Kind: resource.KindWorkerTemplate, Name: "review-worker",
			},
			Inputs: map[string]string{},
			Alias:  "reviewer-42",
		}),
	}
	plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	plan.CanonicalManifest = expertCanonicalJSON(t, manifest)
	plan.DraftHash = expertDigestJSON(t, plan.CanonicalManifest)
	plan.ArtifactKind = resource.KindWorker + "Apply"
	plan.ArtifactJSON = expertCanonicalJSON(t, workerplanner.DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 100,
	})
	plan.ArtifactDigest = expertDigestJSON(t, plan.ArtifactJSON)
	plan.ResolvedReferences = []control.ResolvedReference{}
	plan.PlanHash = expertPlanHash(t, plan)
	return plan
}
