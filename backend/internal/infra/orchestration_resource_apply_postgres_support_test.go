package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/migrations"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func orchestrationPostgresRepository(
	t *testing.T,
) (*gorm.DB, *orchestrationResourceRepo) {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agentsmesh:agentsmesh_dev@localhost:10002/agentsmesh?sslmode=disable"
	}
	admin, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Skipf("PostgreSQL is unavailable: %v", err)
	}
	sqlDB, err := admin.DB()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Skipf("PostgreSQL is unavailable: %v", err)
	}
	schema := fmt.Sprintf("orchestration_repo_%d", time.Now().UnixNano())
	require.NoError(t, admin.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error)
	require.NoError(t, admin.Exec(`CREATE SCHEMA `+schema).Error)
	t.Cleanup(func() {
		_ = admin.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`).Error
	})
	db, err := gorm.Open(
		postgres.Open(orchestrationPostgresDSN(dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(orchestrationPostgresFixtures).Error)
	up, err := migrations.FS.ReadFile("000211_orchestration_resources.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
	integrity, err := migrations.FS.ReadFile(
		"000212_orchestration_resource_integrity.up.sql",
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(integrity)).Error)
	return db, NewOrchestrationResourceRepository(db)
}

func orchestrationPostgresDSN(dsn, schema string) string {
	if strings.Contains(dsn, "://") {
		parsed, err := url.Parse(dsn)
		if err == nil {
			query := parsed.Query()
			query.Set("search_path", schema+",public")
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
	}
	return dsn + " search_path=" + schema + ",public"
}

func orchestrationCreateMutation(
	t *testing.T,
	state orchestrationservice.LockedApplyState,
) orchestrationservice.ApplyMutation {
	t.Helper()
	appliedAt := state.AppliedAt
	manifest := orchestrationStoredManifest(
		t,
		state.Plan,
		state.ResultIdentity,
		1,
		1,
	)
	spec, err := orchestrationcontrol.CanonicalJSONObject(manifest.Spec)
	require.NoError(t, err)
	manifestJSON, err := orchestrationcontrol.CanonicalJSONObject(manifest)
	require.NoError(t, err)
	digest, err := orchestrationcontrol.DigestCanonicalJSON(manifestJSON)
	require.NoError(t, err)
	head := orchestrationcontrol.ResourceHead{
		ID: state.ResultResourceID, OrganizationID: state.Plan.Scope.OrganizationID,
		Identity: state.ResultIdentity, DisplayName: manifest.Metadata.DisplayName,
		Labels: manifest.Metadata.Labels, Status: manifest.Status,
		Revision: 1, Generation: 1, ResourceVersion: 1,
		CreatedByID: state.Plan.ActorID, UpdatedByID: state.Plan.ActorID,
		CreatedAt: appliedAt, UpdatedAt: appliedAt,
	}
	return orchestrationservice.ApplyMutation{
		ArtifactDigest: state.Plan.ArtifactDigest,
		Head:           head,
		Revision: orchestrationcontrol.ResourceRevision{
			OrganizationID: state.Plan.Scope.OrganizationID,
			ResourceID:     state.ResultResourceID, Identity: state.ResultIdentity,
			Revision: 1, Generation: 1, ResourceVersion: 1,
			CanonicalManifest: manifestJSON, CanonicalSpec: spec,
			ResolvedReferences: state.Plan.ResolvedReferences, Digest: digest,
			ActorID: state.Plan.ActorID, CreatedAt: appliedAt,
		},
	}
}

func orchestrationStoredManifest(
	t *testing.T,
	plan orchestrationcontrol.Plan,
	identity orchestrationcontrol.ResourceIdentity,
	resourceVersion, generation int64,
) orchestrationresource.Manifest {
	t.Helper()
	var manifest orchestrationresource.Manifest
	require.NoError(t, json.Unmarshal(plan.CanonicalManifest, &manifest))
	manifest.Metadata.UID = identity.UID
	manifest.Metadata.ResourceVersion = fmt.Sprint(resourceVersion)
	manifest.Metadata.Generation = generation
	manifest.Status = json.RawMessage(`{}`)
	return manifest
}

func orchestrationApplyInitialResource(
	t *testing.T,
	repo orchestrationservice.Repository,
) orchestrationcontrol.ResourceHead {
	return orchestrationApplyInitialResourceForActor(t, repo, 7)
}

func orchestrationApplyInitialResourceForActor(
	t *testing.T,
	repo orchestrationservice.Repository,
	actorID int64,
) orchestrationcontrol.ResourceHead {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	plan.ActorID = actorID
	plan.Scope.ActorID = actorID
	var err error
	plan.PlanHash, err = orchestrationcontrol.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	head, err := repo.RunApplyTransaction(
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
	require.NoError(t, err)
	return head
}

func orchestrationUpdatePlan(
	t *testing.T,
	head orchestrationcontrol.ResourceHead,
	mode string,
) orchestrationcontrol.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	plan.ID = "77777777-7777-4777-8777-777777777777"
	plan.Operation = orchestrationcontrol.PlanOperationUpdate
	plan.TargetResourceID = head.ID
	plan.BaseUID = head.Identity.UID
	plan.BaseResourceVersion = head.ResourceVersion
	plan.CreatedAt = head.UpdatedAt
	plan.ExpiresAt = plan.CreatedAt.Add(5 * time.Minute)
	manifest, err := orchestrationcontrol.CanonicalJSONObject(
		orchestrationresource.Manifest{
			TypeMeta: head.Identity.TypeMeta,
			Metadata: orchestrationresource.Metadata{
				Name: head.Identity.Name, Namespace: head.Identity.Namespace,
			},
			Spec: json.RawMessage(
				fmt.Sprintf(`{"mode":%q,"modelBindingRef":{"kind":"ModelBinding","name":"coding-primary"}}`, mode),
			),
		},
	)
	require.NoError(t, err)
	plan.CanonicalManifest = manifest
	plan.DraftHash, err = orchestrationcontrol.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	plan.PlanHash, err = orchestrationcontrol.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}

func orchestrationUpdateMutation(
	t *testing.T,
	state orchestrationservice.LockedApplyState,
) orchestrationservice.ApplyMutation {
	t.Helper()
	require.NotNil(t, state.Head)
	require.NotNil(t, state.CurrentRevision)
	appliedAt := state.AppliedAt
	head := *state.Head
	head.Revision++
	head.Generation++
	head.ResourceVersion++
	head.UpdatedByID = state.Plan.ActorID
	head.UpdatedAt = appliedAt
	manifest := orchestrationStoredManifest(
		t,
		state.Plan,
		state.ResultIdentity,
		head.ResourceVersion,
		head.Generation,
	)
	manifestJSON, err := orchestrationcontrol.CanonicalJSONObject(manifest)
	require.NoError(t, err)
	spec, err := orchestrationcontrol.CanonicalJSONObject(manifest.Spec)
	require.NoError(t, err)
	digest, err := orchestrationcontrol.DigestCanonicalJSON(manifestJSON)
	require.NoError(t, err)
	head.DisplayName = manifest.Metadata.DisplayName
	head.Labels = manifest.Metadata.Labels
	head.Status = manifest.Status
	return orchestrationservice.ApplyMutation{
		ArtifactDigest: state.Plan.ArtifactDigest,
		Head:           head,
		Revision: orchestrationcontrol.ResourceRevision{
			OrganizationID: state.Plan.Scope.OrganizationID,
			ResourceID:     head.ID, Identity: head.Identity,
			Revision: head.Revision, Generation: head.Generation,
			ResourceVersion:   head.ResourceVersion,
			CanonicalManifest: manifestJSON, CanonicalSpec: spec,
			ResolvedReferences: state.Plan.ResolvedReferences, Digest: digest,
			ActorID: state.Plan.ActorID, CreatedAt: appliedAt,
		},
	}
}

func orchestrationApplyTestCreatePlan(t *testing.T) orchestrationcontrol.Plan {
	t.Helper()
	plan := orchestrationTestCreatePlan(t)
	plan.CreatedAt = time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	plan.ExpiresAt = plan.CreatedAt.Add(10 * time.Minute)
	plan.ResolvedReferences = []orchestrationcontrol.ResolvedReference{{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "ModelBinding",
		},
		Namespace: plan.Scope.OrganizationSlug,
		Name:      "coding-primary",
		UID:       "33333333-3333-4333-8333-333333333333",
		Revision:  1,
		Digest:    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}
	var err error
	plan.PlanHash, err = orchestrationcontrol.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}

const orchestrationPostgresFixtures = `
CREATE TABLE organizations (
	id BIGINT PRIMARY KEY,
	slug VARCHAR(100) NOT NULL UNIQUE
);
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE organization_members (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,
	role VARCHAR(50) NOT NULL,
	joined_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
	UNIQUE (organization_id, user_id)
);
CREATE TABLE worker_spec_snapshots (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	version BIGINT NOT NULL,
	spec_json JSONB NOT NULL,
	summary_json JSONB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
	UNIQUE (organization_id, id)
);
INSERT INTO organizations (id, slug) VALUES (42, 'team-alpha'), (99, 'team-beta');
INSERT INTO users (id) VALUES (7), (8);
INSERT INTO organization_members (organization_id, user_id, role)
VALUES (42, 7, 'owner');
`
