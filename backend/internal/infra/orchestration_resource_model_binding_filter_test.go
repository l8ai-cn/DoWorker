package infra

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestListResourcesFiltersModelBindingsByProtocol(t *testing.T) {
	db, repository := orchestrationRepositoryForTest(t)
	scope := modelBindingScope(t, db)
	insertModelBinding(t, db, scope, scope.ActorID, 201, "openai-chat", 11, 21, "openai")
	insertModelBinding(t, db, scope, scope.ActorID, 202, "minimax-chat", 12, 22, "minimax")

	page, err := repository.ListResources(
		context.Background(),
		scope,
		service.ResourceListFilter{
			Kind: resource.KindModelBinding, Limit: 50,
			ModelBinding: &service.ModelBindingReferenceFilter{
				WorkerType:       slugkit.MustNewForTest("minimax-cli"),
				ProtocolAdapters: []string{"minimax"},
			},
		},
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "minimax-chat", page.Items[0].Identity.Name.String())
	assert.EqualValues(t, 1, page.Total)
}

func TestListResourcesIgnoresBrokenModelBindingsOutsideWorkerCandidates(t *testing.T) {
	db, repository := orchestrationRepositoryForTest(t)
	scope := modelBindingScope(t, db)
	insertModelBinding(t, db, scope, scope.ActorID, 201, "missing-model", 99, 0, "")
	insertModelBinding(t, db, scope, scope.ActorID, 202, "minimax-chat", 12, 22, "minimax")

	page, err := repository.ListResources(
		context.Background(),
		scope,
		service.ResourceListFilter{
			Kind: resource.KindModelBinding, Limit: 50,
			ModelBinding: &service.ModelBindingReferenceFilter{
				WorkerType:       slugkit.MustNewForTest("minimax-cli"),
				ProtocolAdapters: []string{"minimax"},
			},
		},
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "minimax-chat", page.Items[0].Identity.Name.String())
}

func TestListResourcesOnlyReturnsSelectableModelBindings(t *testing.T) {
	tests := []struct {
		name    string
		ownerID func(*testing.T, *gorm.DB, control.Scope) int64
		mutate  func(*testing.T, *gorm.DB)
	}{
		{
			name: "another user",
			ownerID: func(t *testing.T, db *gorm.DB, _ control.Scope) int64 {
				return testkit.CreateUser(
					t,
					db,
					"other-model-owner@example.test",
					"other-model-owner",
				)
			},
		},
		{
			name: "disabled connection",
			mutate: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(
					"UPDATE provider_connections SET is_enabled = false WHERE id = 21",
				).Error)
			},
		},
		{
			name: "unchecked model",
			mutate: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(
					"UPDATE model_resources SET status = 'unchecked' WHERE id = 11",
				).Error)
			},
		},
		{
			name: "video only",
			mutate: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(
					"UPDATE model_resources SET modalities = '[\"video\"]' WHERE id = 11",
				).Error)
			},
		},
		{
			name: "without text generation",
			mutate: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(
					"UPDATE model_resources SET capabilities = '[\"video-generation\"]' WHERE id = 11",
				).Error)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, repository := orchestrationRepositoryForTest(t)
			scope := modelBindingScope(t, db)
			ownerID := scope.ActorID
			if test.ownerID != nil {
				ownerID = test.ownerID(t, db, scope)
			}
			insertModelBinding(
				t,
				db,
				scope,
				ownerID,
				201,
				"minimax-chat",
				11,
				21,
				"minimax",
			)
			if test.mutate != nil {
				test.mutate(t, db)
			}

			page, err := repository.ListResources(
				context.Background(),
				scope,
				service.ResourceListFilter{
					Kind: resource.KindModelBinding, Limit: 50,
					ModelBinding: &service.ModelBindingReferenceFilter{
						WorkerType:       slugkit.MustNewForTest("minimax-cli"),
						ProtocolAdapters: []string{"minimax"},
					},
				},
			)

			require.NoError(t, err)
			assert.Empty(t, page.Items)
			assert.Zero(t, page.Total)
		})
	}
}

func insertModelBinding(
	t *testing.T,
	db *gorm.DB,
	scope control.Scope,
	ownerID int64,
	resourceID int64,
	name string,
	modelID int64,
	connectionID int64,
	providerKey string,
) {
	t.Helper()
	if connectionID > 0 {
		require.NoError(t, db.Exec(
			`INSERT INTO provider_connections (
				id, owner_scope, owner_id, identifier, provider_key, name,
				configured_fields, status, is_enabled, created_by
			) VALUES (?, 'user', ?, ?, ?, ?, '[]', 'valid', true, ?)`,
			connectionID, ownerID, name+"-provider", providerKey, name, scope.ActorID,
		).Error)
		require.NoError(t, db.Exec(
			`INSERT INTO model_resources (
				id, provider_connection_id, identifier, model_id, display_name,
				modalities, capabilities, status, is_enabled
			) VALUES (?, ?, ?, ?, ?, '["chat"]', '["text-generation"]', 'valid', true)`,
			modelID, connectionID, name+"-model", name, name,
		).Error)
	}
	head := orchestrationTestHead()
	head.ID = resourceID
	head.OrganizationID = scope.OrganizationID
	head.Identity.Namespace = scope.OrganizationSlug
	head.CreatedByID = scope.ActorID
	head.UpdatedByID = scope.ActorID
	head.Identity.ResourceTarget.Kind = resource.KindModelBinding
	head.Identity.Name = slugkit.MustNewForTest(name)
	revision := orchestrationTestRevision(t, head)
	spec, err := control.CanonicalJSONObject(map[string]any{"resourceId": modelID})
	require.NoError(t, err)
	manifest, err := control.CanonicalJSONObject(resource.Manifest{
		TypeMeta: head.Identity.TypeMeta,
		Metadata: resource.Metadata{
			Name: head.Identity.Name, Namespace: head.Identity.Namespace,
			DisplayName: head.DisplayName, Labels: head.Labels,
			UID: head.Identity.UID, ResourceVersion: "1", Generation: 1,
		},
		Spec: json.RawMessage(spec), Status: head.Status,
	})
	require.NoError(t, err)
	digest, err := control.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	revision.CanonicalManifest = manifest
	revision.CanonicalSpec = spec
	revision.Digest = digest
	insertOrchestrationHead(t, db, head)
	record, err := orchestrationRevisionRecordFromDomain(revision, scope)
	require.NoError(t, err)
	require.NoError(t, db.Create(&record).Error)
}

func modelBindingScope(t *testing.T, db *gorm.DB) control.Scope {
	t.Helper()
	actorID := testkit.CreateUser(
		t,
		db,
		"model-binding-owner@example.test",
		"model-binding-owner",
	)
	scope := orchestrationTestScope()
	scope.ActorID = actorID
	return scope
}
