package infra

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestExpertApplyTransactionCreatesPinnedProjectionAndReplays(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	applyOrchestrationDomainLinkFixtures(t, db)
	plan := orchestrationExpertApplyPlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	builder := func(state controlservice.LockedApplyState) (
		workerplanner.ExpertApplyMutation,
		error,
	) {
		mutation := orchestrationCreateMutation(t, state)
		mutation.Revision.WorkerSpecSnapshotID = 100
		return workerplanner.ExpertApplyMutation{
			ApplyMutation: mutation,
			Projection: workerplanner.ExpertApplyProjection{
				Name: "Review Expert", Description: "Reviews changes",
				Category: "engineering", ReleaseNotes: "Initial revision",
				Prompt: "Review carefully", WorkerSpecSnapshotID: 100,
			},
		}, nil
	}

	first, err := repo.RunExpertApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)
	replayed, err := repo.RunExpertApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)

	assert.Equal(t, first, replayed)
	assert.Equal(t, int64(100), first.WorkerSpecSnapshotID)
	var count int64
	require.NoError(t, db.Table("experts").Count(&count).Error)
	assert.Equal(t, int64(1), count)
	var resourceID, revision, snapshotID int64
	require.NoError(t, db.Table("experts").
		Select("orchestration_resource_id, orchestration_resource_revision, worker_spec_snapshot_id").
		Row().Scan(&resourceID, &revision, &snapshotID))
	assert.Equal(t, first.Head.ID, resourceID)
	assert.Equal(t, first.Head.Revision, revision)
	assert.Equal(t, first.WorkerSpecSnapshotID, snapshotID)
}

func applyOrchestrationDomainLinkFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
INSERT INTO worker_spec_snapshots
	(id, organization_id, version, spec_json, summary_json)
VALUES (100, 42, 1, '{}', '{}');
CREATE TABLE experts (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	slug VARCHAR(100) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description TEXT,
	agent_slug VARCHAR(100) NOT NULL,
	prompt TEXT,
	interaction_mode VARCHAR(20) NOT NULL DEFAULT 'acp',
	automation_level VARCHAR(20) NOT NULL DEFAULT 'autonomous',
	worker_spec_snapshot_id BIGINT,
	metadata JSONB NOT NULL DEFAULT '{}',
	created_by_id BIGINT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_experts_org_slug ON experts(organization_id, slug);
CREATE TABLE workflows (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	name VARCHAR(255) NOT NULL,
	slug VARCHAR(100) NOT NULL,
	description TEXT,
	agent_slug VARCHAR(100),
	permission_mode VARCHAR(50) NOT NULL DEFAULT 'bypassPermissions',
	prompt_template TEXT NOT NULL,
	repository_id BIGINT,
	runner_id BIGINT,
	branch_name VARCHAR(255),
	ticket_id BIGINT,
	model_resource_id BIGINT,
	used_env_bundles TEXT[] NOT NULL DEFAULT '{}',
	config_overrides JSONB DEFAULT '{}',
	prompt_variables JSONB DEFAULT '{}',
	execution_mode VARCHAR(20) NOT NULL DEFAULT 'autopilot',
	cron_expression VARCHAR(100),
	autopilot_config JSONB NOT NULL DEFAULT '{}',
	callback_url VARCHAR(500),
	status VARCHAR(20) NOT NULL DEFAULT 'enabled',
	sandbox_strategy VARCHAR(20) NOT NULL DEFAULT 'persistent',
	session_persistence BOOLEAN NOT NULL DEFAULT true,
	concurrency_policy VARCHAR(20) NOT NULL DEFAULT 'skip',
	max_concurrent_runs INT NOT NULL DEFAULT 1,
	max_retained_runs INT NOT NULL DEFAULT 0,
	timeout_minutes INT NOT NULL DEFAULT 60,
	idle_timeout_sec INT NOT NULL DEFAULT 30,
	sandbox_path VARCHAR(500),
	last_pod_key VARCHAR(100),
	created_by_id BIGINT NOT NULL,
	total_runs INT NOT NULL DEFAULT 0,
	successful_runs INT NOT NULL DEFAULT 0,
	failed_runs INT NOT NULL DEFAULT 0,
	last_run_at TIMESTAMPTZ,
	next_run_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (organization_id, slug)
);
CREATE TABLE workflow_runs (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	workflow_id BIGINT NOT NULL,
	run_number INT NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	pod_key VARCHAR(100),
	autopilot_controller_key VARCHAR(100),
	trigger_type VARCHAR(20) NOT NULL,
	trigger_source VARCHAR(255),
	trigger_params JSONB DEFAULT '{}',
	resolved_prompt TEXT,
	started_at TIMESTAMPTZ,
	finished_at TIMESTAMPTZ,
	duration_sec INT,
	exit_summary TEXT,
	error_message TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (workflow_id, run_number)
);`).Error)
	up, err := migrations.FS.ReadFile(
		"000215_orchestration_domain_links.up.sql",
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
}

func orchestrationExpertApplyPlan(t *testing.T) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindExpert,
		},
		Metadata: resource.Metadata{
			Name: "review-expert", Namespace: "team-alpha",
			DisplayName: "Review Expert", Labels: map[string]string{},
		},
		Spec: expertCanonicalJSON(t, resource.ExpertResourceSpec{
			WorkerTemplateRef: resource.Reference{
				Kind: resource.KindWorkerTemplate, Name: "review-worker",
			},
			Description: "Reviews changes", Category: "engineering",
			ReleaseNotes: "Initial revision",
		}),
	}
	plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	plan.CanonicalManifest = expertCanonicalJSON(t, manifest)
	plan.DraftHash = expertDigestJSON(t, plan.CanonicalManifest)
	plan.ArtifactKind = resource.KindExpert + "Apply"
	plan.ArtifactJSON = expertCanonicalJSON(t, workerplanner.DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 100,
	})
	plan.ArtifactDigest = expertDigestJSON(t, plan.ArtifactJSON)
	plan.ResolvedReferences = []control.ResolvedReference{}
	plan.PlanHash = expertPlanHash(t, plan)
	return plan
}

func expertCanonicalJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	canonical, err := control.CanonicalJSONObject(value)
	require.NoError(t, err)
	return canonical
}

func expertDigestJSON(t *testing.T, value json.RawMessage) string {
	t.Helper()
	digest, err := control.DigestCanonicalJSON(value)
	require.NoError(t, err)
	return digest
}

func expertPlanHash(t *testing.T, plan control.Plan) string {
	t.Helper()
	hash, err := control.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return hash
}
