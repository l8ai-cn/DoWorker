package infra

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestListResourcesFiltersEnvironmentBundlesByPurposeAndWorker(
	t *testing.T,
) {
	db, repo := orchestrationRepositoryForTest(t)
	insertReferenceBundle(t, db, 501, "user", 7, "do-agent", "config", true)
	insertReferenceBundle(t, db, 502, "user", 7, "do-agent", "credential", true)
	insertReferenceBundle(t, db, 503, "user", 7, "other-agent", "config", true)
	insertReferenceBundle(t, db, 504, "user", 8, "do-agent", "config", true)
	insertReferenceBundle(t, db, 505, "org", 42, "", "config", true)
	insertReferenceBundle(t, db, 506, "org", 99, "", "config", true)
	insertReferenceBundle(t, db, 507, "user", 7, "do-agent", "runtime", false)
	insertReferenceBundle(t, db, 508, "user", 7, "do-agent", "shared", true)
	for index, bundleID := range []int64{501, 502, 503, 504, 505, 506, 507, 508} {
		insertEnvironmentBundleResource(
			t,
			db,
			int64(301+index),
			fmt.Sprintf("bundle-%d", bundleID),
			bundleID,
		)
	}

	config := listEnvironmentBundleReferences(
		t,
		repo,
		service.EnvironmentBundlePurposeConfig,
	)
	assert.Equal(t, []string{"bundle-501", "bundle-505"}, config)

	runtime := listEnvironmentBundleReferences(
		t,
		repo,
		service.EnvironmentBundlePurposeRuntime,
	)
	assert.Equal(t, []string{"bundle-508"}, runtime)
}

func TestEnvironmentBundlePurposeKindsMatchWorkerCompilation(t *testing.T) {
	runtime, err := environmentBundlePurposeKinds(
		service.EnvironmentBundlePurposeRuntime,
	)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"runtime", "shared"}, runtime)

	config, err := environmentBundlePurposeKinds(
		service.EnvironmentBundlePurposeConfig,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"config"}, config)

	credential, err := environmentBundlePurposeKinds(
		service.EnvironmentBundlePurposeCredential,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"credential"}, credential)
}

func TestEnvironmentBundleFilterRejectsUnsupportedDatabaseDialect(t *testing.T) {
	_, err := environmentBundleEntityIDExpression("mysql")
	require.ErrorIs(t, err, service.ErrUnavailable)
}

func TestEnvironmentBundleIDValidityPredicateRejectsUnsupportedDialect(
	t *testing.T,
) {
	_, err := environmentBundleIDValidityPredicate("mysql")
	require.ErrorIs(t, err, service.ErrUnavailable)
}

func TestListResourcesRejectsBrokenEnvironmentBundleBinding(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	insertEnvironmentBundleResource(t, db, 301, "missing-bundle", 999)

	_, err := repo.ListResources(
		t.Context(),
		orchestrationTestScope(),
		service.ResourceListFilter{
			Kind:   orchestrationresource.KindEnvironmentBundle,
			Limit:  100,
			Offset: 0,
			EnvironmentBundle: &service.EnvironmentBundleReferenceFilter{
				Purpose:    service.EnvironmentBundlePurposeConfig,
				WorkerType: slugkit.Slug("do-agent"),
			},
		},
	)

	require.ErrorIs(t, err, service.ErrUnavailable)
}

func TestListResourcesFiltersEnvironmentBundleDataKeys(t *testing.T) {
	db, repo := orchestrationRepositoryForTest(t)
	insertReferenceBundleData(
		t, db, 601, "user", 7, "cursor-cli", "runtime", true,
		`{"CURSOR_MODEL":"managed"}`,
	)
	insertReferenceBundleData(
		t, db, 602, "user", 7, "cursor-cli", "runtime", true,
		`{"LOG_LEVEL":"debug"}`,
	)
	insertReferenceBundleData(
		t, db, 603, "user", 7, "cursor-cli", "credential", true,
		`{"CURSOR_API_KEY":"encrypted"}`,
	)
	insertReferenceBundleData(
		t, db, 604, "user", 7, "cursor-cli", "credential", true,
		`{"OTHER_KEY":"encrypted"}`,
	)
	for index, bundleID := range []int64{601, 602, 603, 604} {
		insertEnvironmentBundleResource(
			t, db, int64(401+index),
			fmt.Sprintf("bundle-%d", bundleID), bundleID,
		)
	}

	runtime := listEnvironmentBundleReferencesWithFilter(
		t,
		repo,
		service.EnvironmentBundleReferenceFilter{
			Purpose:            service.EnvironmentBundlePurposeRuntime,
			WorkerType:         slugkit.Slug("cursor-cli"),
			ModelManagedFields: []string{"CURSOR_MODEL"},
		},
	)
	assert.Equal(t, []string{"bundle-602"}, runtime)

	credential := listEnvironmentBundleReferencesWithFilter(
		t,
		repo,
		service.EnvironmentBundleReferenceFilter{
			Purpose:    service.EnvironmentBundlePurposeCredential,
			WorkerType: slugkit.Slug("cursor-cli"),
			TargetName: "CURSOR_API_KEY",
		},
	)
	assert.Equal(t, []string{"bundle-603"}, credential)
}

func listEnvironmentBundleReferences(
	t *testing.T,
	repo service.Repository,
	purpose service.EnvironmentBundlePurpose,
) []string {
	t.Helper()
	return listEnvironmentBundleReferencesWithFilter(
		t,
		repo,
		service.EnvironmentBundleReferenceFilter{
			Purpose: purpose, WorkerType: slugkit.Slug("do-agent"),
		},
	)
}

func listEnvironmentBundleReferencesWithFilter(
	t *testing.T,
	repo service.Repository,
	filter service.EnvironmentBundleReferenceFilter,
) []string {
	t.Helper()
	page, err := repo.ListResources(
		t.Context(),
		orchestrationTestScope(),
		service.ResourceListFilter{
			Kind:              orchestrationresource.KindEnvironmentBundle,
			Limit:             100,
			Offset:            0,
			EnvironmentBundle: &filter,
		},
	)
	require.NoError(t, err)
	require.Equal(t, int64(len(page.Items)), page.Total)
	names := make([]string, len(page.Items))
	for index, item := range page.Items {
		names[index] = item.Identity.Name.String()
	}
	return names
}

func insertReferenceBundle(
	t *testing.T,
	db *gorm.DB,
	id int64,
	ownerScope string,
	ownerID int64,
	agentSlug string,
	kind string,
	active bool,
) {
	t.Helper()
	insertReferenceBundleData(
		t, db, id, ownerScope, ownerID, agentSlug, kind, active, `{}`,
	)
}

func insertReferenceBundleData(
	t *testing.T,
	db *gorm.DB,
	id int64,
	ownerScope string,
	ownerID int64,
	agentSlug string,
	kind string,
	active bool,
	data string,
) {
	t.Helper()
	var workerType any
	if agentSlug != "" {
		workerType = agentSlug
	}
	require.NoError(t, db.Exec(`
INSERT INTO env_bundles (
  id, owner_scope, owner_id, agent_slug, name, kind, kind_primary, data,
  is_active, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, ownerScope, ownerID, workerType, fmt.Sprintf("source-%d", id), kind,
		false, data, active, orchestrationRepoTestTime, orchestrationRepoTestTime,
	).Error)
}

func insertEnvironmentBundleResource(
	t *testing.T,
	db *gorm.DB,
	resourceID int64,
	name string,
	bundleID int64,
) {
	t.Helper()
	head := orchestrationTestHead()
	head.ID = resourceID
	head.Identity.Kind = orchestrationresource.KindEnvironmentBundle
	head.Identity.Name = slugkit.Slug(name)
	head.Identity.UID = fmt.Sprintf(
		"00000000-0000-4000-8000-%012d",
		resourceID,
	)
	head.DisplayName = name
	insertOrchestrationHead(t, db, head)

	spec, err := orchestrationcontrol.CanonicalJSONObject(map[string]any{
		"environmentBundleId": bundleID,
	})
	require.NoError(t, err)
	manifest, err := orchestrationcontrol.CanonicalJSONObject(
		orchestrationresource.Manifest{
			TypeMeta: head.Identity.TypeMeta,
			Metadata: orchestrationresource.Metadata{
				Name: head.Identity.Name, Namespace: head.Identity.Namespace,
				DisplayName: head.DisplayName, Labels: head.Labels,
				UID: head.Identity.UID, ResourceVersion: "1", Generation: 1,
			},
			Spec: json.RawMessage(spec), Status: head.Status,
		},
	)
	require.NoError(t, err)
	digest, err := orchestrationcontrol.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	insertOrchestrationRevision(t, db, orchestrationcontrol.ResourceRevision{
		OrganizationID: head.OrganizationID, ResourceID: head.ID,
		Identity: head.Identity, Revision: 1, Generation: 1, ResourceVersion: 1,
		CanonicalManifest: manifest, CanonicalSpec: spec,
		ResolvedReferences: []orchestrationcontrol.ResolvedReference{},
		Digest:             digest, ActorID: 7, CreatedAt: orchestrationRepoTestTime,
	})
}
