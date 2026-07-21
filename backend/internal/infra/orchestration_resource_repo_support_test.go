package infra

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	orchestrationservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var orchestrationRepoTestTime = time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)

func orchestrationRepositoryForTest(
	t *testing.T,
) (*gorm.DB, orchestrationservice.Repository) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	for _, statement := range orchestrationRepositoryTestDDL {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db, NewOrchestrationResourceRepository(db)
}

func orchestrationTestScope() orchestrationcontrol.Scope {
	return orchestrationcontrol.Scope{
		OrganizationID:   42,
		OrganizationSlug: slugkit.MustNewForTest("team-alpha"),
		ActorID:          7,
	}
}

func orchestrationTestTarget() orchestrationcontrol.ResourceTarget {
	return orchestrationcontrol.ResourceTarget{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest("worker-one"),
	}
}

func orchestrationTestHead() orchestrationcontrol.ResourceHead {
	return orchestrationcontrol.ResourceHead{
		ID:             101,
		OrganizationID: 42,
		Identity: orchestrationcontrol.ResourceIdentity{
			ResourceTarget: orchestrationTestTarget(),
			UID:            "22222222-2222-4222-8222-222222222222",
		},
		DisplayName:     "Worker One",
		Labels:          map[string]string{"role": "builder"},
		Status:          json.RawMessage(`{"ready":true}`),
		Revision:        1,
		Generation:      1,
		ResourceVersion: 1,
		CreatedByID:     7,
		UpdatedByID:     7,
		CreatedAt:       orchestrationRepoTestTime,
		UpdatedAt:       orchestrationRepoTestTime,
	}
}

func orchestrationTestRevision(
	t *testing.T,
	head orchestrationcontrol.ResourceHead,
) orchestrationcontrol.ResourceRevision {
	t.Helper()
	manifest, err := orchestrationcontrol.CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: head.Identity.TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name:            head.Identity.Name,
			Namespace:       head.Identity.Namespace,
			DisplayName:     head.DisplayName,
			Labels:          head.Labels,
			UID:             head.Identity.UID,
			ResourceVersion: "1",
			Generation:      1,
		},
		Spec:   json.RawMessage(`{"model":"coding-primary"}`),
		Status: head.Status,
	})
	require.NoError(t, err)
	spec, err := orchestrationcontrol.CanonicalJSONObject(
		json.RawMessage(`{"model":"coding-primary"}`),
	)
	require.NoError(t, err)
	digest, err := orchestrationcontrol.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	return orchestrationcontrol.ResourceRevision{
		OrganizationID:     42,
		ResourceID:         head.ID,
		Identity:           head.Identity,
		Revision:           1,
		Generation:         1,
		ResourceVersion:    1,
		CanonicalManifest:  manifest,
		CanonicalSpec:      spec,
		ResolvedReferences: []orchestrationcontrol.ResolvedReference{},
		Digest:             digest,
		ActorID:            7,
		CreatedAt:          orchestrationRepoTestTime,
	}
}

func orchestrationTestCreatePlan(t *testing.T) orchestrationcontrol.Plan {
	t.Helper()
	target := orchestrationTestTarget()
	manifest, err := orchestrationcontrol.CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: target.TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name: target.Name, Namespace: target.Namespace,
		},
		Spec: json.RawMessage(`{"modelBindingRef":{"kind":"ModelBinding","name":"coding-primary"}}`),
	})
	require.NoError(t, err)
	artifact, err := orchestrationcontrol.CanonicalJSONObject(
		map[string]any{"workerSpecVersion": 1},
	)
	require.NoError(t, err)
	plan := orchestrationcontrol.Plan{
		ID:    "11111111-1111-4111-8111-111111111111",
		Scope: orchestrationTestScope(), ActorID: 7,
		Operation: orchestrationcontrol.PlanOperationCreate, Target: target,
		CanonicalManifest:  manifest,
		ResolvedReferences: []orchestrationcontrol.ResolvedReference{},
		SemanticChanges:    []orchestrationcontrol.SemanticChange{},
		Issues:             []orchestrationcontrol.PlanIssue{},
		ArtifactKind:       "WorkerSpec", ArtifactJSON: artifact,
		OptionsRevision: "runtime-catalog-1", CreatedAt: orchestrationRepoTestTime,
		ExpiresAt: orchestrationRepoTestTime.Add(5 * time.Minute),
		Status:    orchestrationcontrol.PlanStatusPending,
	}
	plan.DraftHash, err = orchestrationcontrol.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	plan.ArtifactDigest, err = orchestrationcontrol.DigestCanonicalJSON(artifact)
	require.NoError(t, err)
	plan.PlanHash, err = orchestrationcontrol.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}

func insertOrchestrationHead(
	t *testing.T,
	db *gorm.DB,
	head orchestrationcontrol.ResourceHead,
) {
	t.Helper()
	scope := orchestrationcontrol.Scope{
		OrganizationID:   head.OrganizationID,
		OrganizationSlug: head.Identity.Namespace,
		ActorID:          head.CreatedByID,
	}
	record, err := orchestrationResourceRecordFromDomain(head, scope)
	require.NoError(t, err)
	require.NoError(t, db.Create(&record).Error)
}

func insertOrchestrationRevision(
	t *testing.T,
	db *gorm.DB,
	revision orchestrationcontrol.ResourceRevision,
) {
	t.Helper()
	record, err := orchestrationRevisionRecordFromDomain(revision, orchestrationTestScope())
	require.NoError(t, err)
	require.NoError(t, db.Create(&record).Error)
}

var orchestrationRepositoryTestDDL = []string{
	`CREATE TABLE orchestration_resources (
		id INTEGER PRIMARY KEY, organization_id INTEGER NOT NULL, uid TEXT NOT NULL,
		api_version TEXT NOT NULL, kind TEXT NOT NULL, namespace TEXT NOT NULL,
		name TEXT NOT NULL, display_name TEXT NOT NULL, labels BLOB NOT NULL,
		status BLOB NOT NULL, generation INTEGER NOT NULL, resource_version INTEGER NOT NULL,
		active_revision INTEGER NOT NULL, created_by_id INTEGER NOT NULL,
		updated_by_id INTEGER NOT NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL,
		UNIQUE(organization_id, api_version, kind, namespace, name)
	)`,
	`CREATE TABLE orchestration_resource_revisions (
		id INTEGER PRIMARY KEY AUTOINCREMENT, organization_id INTEGER NOT NULL,
		resource_id INTEGER NOT NULL, revision INTEGER NOT NULL, generation INTEGER NOT NULL,
		resource_version INTEGER NOT NULL, canonical_manifest BLOB NOT NULL,
		canonical_spec BLOB NOT NULL, resolved_refs BLOB NOT NULL, digest TEXT NOT NULL,
		worker_spec_snapshot_id INTEGER, actor_id INTEGER NOT NULL, created_at DATETIME NOT NULL,
		UNIQUE(resource_id, revision)
	)`,
	`CREATE TABLE orchestration_resource_plans (
		id TEXT PRIMARY KEY, organization_id INTEGER NOT NULL, actor_id INTEGER NOT NULL,
		target_resource_id INTEGER, target_api_version TEXT NOT NULL, target_kind TEXT NOT NULL,
		target_namespace TEXT NOT NULL, target_name TEXT NOT NULL, operation TEXT NOT NULL,
		base_head_uid TEXT, base_resource_version INTEGER, draft_hash TEXT NOT NULL,
		plan_hash TEXT NOT NULL, canonical_manifest BLOB NOT NULL, resolved_refs BLOB NOT NULL,
		semantic_diff BLOB NOT NULL, issues BLOB NOT NULL, artifact_kind TEXT NOT NULL,
		artifact_json BLOB NOT NULL, artifact_digest TEXT NOT NULL, options_revision TEXT NOT NULL,
		created_at DATETIME NOT NULL, expires_at DATETIME NOT NULL, consumed_at DATETIME,
		consumed_by_id INTEGER, consumption_result TEXT, result_resource_id INTEGER,
		result_resource_uid TEXT, result_resource_version INTEGER, result_revision INTEGER,
		result_json BLOB
	)`,
}
