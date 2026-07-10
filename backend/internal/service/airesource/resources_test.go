package airesource

import (
	"context"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceCRUDRequiresManagementPermission(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeOrg, 10, "org-main", "secret")
	_, err := f.service.CreateResource(context.Background(), actor(3), CreateResourceInput{ConnectionID: connection.ID, Identifier: "member-model", ModelID: "gpt-4.1", DisplayName: "Member model", Modalities: []domain.Modality{domain.ModalityChat}})
	assert.ErrorIs(t, err, ErrForbidden)
	resource, err := f.service.CreateResource(context.Background(), actor(2), CreateResourceInput{ConnectionID: connection.ID, Identifier: "admin-model", ModelID: "gpt-4.1", DisplayName: "Admin model", Modalities: []domain.Modality{domain.ModalityChat}, Capabilities: []domain.Capability{domain.CapabilityTextGeneration}})
	require.NoError(t, err)
	assert.Equal(t, domain.ConnectionStatusValid, resource.Status)
	_, err = f.service.UpdateResource(context.Background(), actor(3), resource.ID, UpdateResourceInput{DisplayName: "Nope", ModelID: resource.ModelID, Modalities: resource.Modalities, Capabilities: resource.Capabilities})
	assert.ErrorIs(t, err, ErrForbidden)
	assert.ErrorIs(t, f.service.DeleteResource(context.Background(), actor(3), resource.ID), ErrForbidden)
}

func TestSetDefaultChecksPermissionAndModality(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeOrg, 10, "org-main", "secret")
	resource := createResource(t, f, connection.ID, "chat-model")
	assert.ErrorIs(t, f.service.SetDefault(context.Background(), actor(3), resource.ID, domain.ModalityChat), ErrForbidden)
	assert.ErrorIs(t, f.service.SetDefault(context.Background(), actor(2), resource.ID, domain.ModalityVideo), ErrIncompatibleModality)
	require.NoError(t, f.service.SetDefault(context.Background(), actor(2), resource.ID, domain.ModalityChat))
	stored := f.repo.resources[resource.ID]
	assert.Equal(t, []domain.Modality{domain.ModalityChat}, stored.DefaultModalities)
}

func TestListOwnerIncludesDisabledAndInvalidResourcesWithoutSecrets(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "model-a")
	stored := f.repo.resources[resource.ID]
	stored.IsEnabled = false
	stored.Status = domain.ConnectionStatusInvalid
	stored.ValidationError = "unavailable"
	views, err := f.service.ListOwnerConnections(context.Background(), actor(1), domain.OwnerScopeUser, 1)
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Len(t, views[0].Resources, 1)
	assert.False(t, views[0].Resources[0].IsEnabled)
	assert.Equal(t, domain.ConnectionStatusInvalid, views[0].Resources[0].Status)
	assert.True(t, views[0].CanManage)
}

func TestListEffectiveValidatesOrganizationMembership(t *testing.T) {
	f := newFixture()
	userConnection := createValidConnection(t, f, domain.OwnerScopeUser, 3, "user-main", "user-secret")
	createResource(t, f, userConnection.ID, "user-model")
	orgConnection := createValidConnection(t, f, domain.OwnerScopeOrg, 10, "org-main", "org-secret")
	createResource(t, f, orgConnection.ID, "org-model")
	views, err := f.service.ListEffective(context.Background(), actor(3), 10, []domain.Modality{domain.ModalityChat})
	require.NoError(t, err)
	assert.Len(t, views, 2)
	_, err = f.service.ListEffective(context.Background(), actor(99), 10, []domain.Modality{domain.ModalityChat})
	assert.ErrorIs(t, err, ErrForbidden)
	f.members.err = errInjected
	_, err = f.service.ListEffective(context.Background(), actor(3), 10, nil)
	assert.ErrorIs(t, err, errInjected)
}

func TestResourceValidationRejectsProviderUnsupportedModality(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	_, err := f.service.CreateResource(context.Background(), actor(1), CreateResourceInput{ConnectionID: connection.ID, Identifier: slugkit.Slug("bad-model"), ModelID: "bad", DisplayName: "Bad", Modalities: []domain.Modality{domain.Modality("telepathy")}})
	assert.ErrorIs(t, err, ErrIncompatibleModality)
}

func TestResourceValidationRejectsCapabilityMismatch(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	_, err := f.service.CreateResource(context.Background(), actor(1), CreateResourceInput{
		ConnectionID: connection.ID, Identifier: "bad-video", ModelID: "bad-video", DisplayName: "Bad video",
		Modalities: []domain.Modality{domain.ModalityVideo}, Capabilities: []domain.Capability{domain.CapabilityTextGeneration},
	})
	assert.ErrorIs(t, err, ErrIncompatibleModality)
}

func TestResourceValidationReturnsTypedInvalidRequestError(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	_, err := f.service.CreateResource(context.Background(), actor(1), CreateResourceInput{
		ConnectionID: connection.ID, Identifier: "incomplete-model",
	})
	assert.ErrorIs(t, err, ErrInvalidRequirements)
}

func TestListEffectiveRejectsInvalidModality(t *testing.T) {
	f := newFixture()
	_, err := f.service.ListEffective(context.Background(), actor(1), 0, []domain.Modality{"telepathy"})
	assert.ErrorIs(t, err, ErrIncompatibleModality)
}

func TestListEffectiveReturnsSelectableAndBlockedResources(t *testing.T) {
	f := newFixture()
	validConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "valid-main", "secret")
	valid := createResource(t, f, validConnection.ID, "valid-model")
	require.NoError(t, f.service.SetDefault(context.Background(), actor(1), valid.ID, domain.ModalityChat))

	disabledConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "disabled-connection", "secret")
	disabledByConnection := createResource(t, f, disabledConnection.ID, "connection-blocked")
	f.repo.connections[disabledConnection.ID].IsEnabled = false
	disabledResourceConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "disabled-resource", "secret")
	disabledResource := createResource(t, f, disabledResourceConnection.ID, "resource-blocked")
	f.repo.resources[disabledResource.ID].IsEnabled = false
	uncheckedConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "unchecked-main", "secret")
	unchecked := createResource(t, f, uncheckedConnection.ID, "unchecked-model")
	f.repo.connections[uncheckedConnection.ID].Status = domain.ConnectionStatusUnchecked
	invalidConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "invalid-main", "secret")
	invalid := createResource(t, f, invalidConnection.ID, "invalid-model")
	f.repo.resources[invalid.ID].Status = domain.ConnectionStatusInvalid
	require.NoError(t, f.repo.SetDefault(context.Background(), invalid.ID, domain.ModalityChat))

	views, err := f.service.ListEffective(context.Background(), actor(1), 0, []domain.Modality{domain.ModalityChat})
	require.NoError(t, err)
	require.Len(t, views, 5)
	byID := map[int64]EffectiveResourceView{}
	for _, view := range views {
		byID[view.Resource.ID] = view
	}
	assert.True(t, byID[valid.ID].Selectable)
	assert.Empty(t, byID[valid.ID].BlockingReason)
	assert.Equal(t, BlockingConnectionDisabled, byID[disabledByConnection.ID].BlockingReason)
	assert.Equal(t, BlockingResourceDisabled, byID[disabledResource.ID].BlockingReason)
	assert.Equal(t, BlockingConnectionUnchecked, byID[unchecked.ID].BlockingReason)
	assert.Equal(t, BlockingResourceInvalid, byID[invalid.ID].BlockingReason)
	assert.Empty(t, byID[invalid.ID].Resource.DefaultModalities)
}

func TestListEffectiveDistinguishesEmptyFromBlocked(t *testing.T) {
	f := newFixture()
	views, err := f.service.ListEffective(context.Background(), actor(1), 0, []domain.Modality{domain.ModalityChat})
	require.NoError(t, err)
	assert.Empty(t, views)
}

func TestOwnerAndEffectiveListsUseOneBulkResourceReadPerOwner(t *testing.T) {
	f := newFixture()
	for _, identifier := range []string{"first-main", "second-main", "third-main"} {
		connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, identifier, "secret")
		createResource(t, f, connection.ID, identifier+"-model")
	}
	f.repo.calls = map[string]int{}
	_, err := f.service.ListOwnerConnections(context.Background(), actor(1), domain.OwnerScopeUser, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, f.repo.calls["ListResourcesByOwner"])
	assert.Zero(t, f.repo.calls["ListResourcesByConnection"])
	f.repo.calls = map[string]int{}
	_, err = f.service.ListEffective(context.Background(), actor(1), 0, []domain.Modality{domain.ModalityChat})
	require.NoError(t, err)
	assert.Equal(t, 1, f.repo.calls["ListResourcesByOwner"])
	assert.Zero(t, f.repo.calls["ListResourcesByConnection"])
}

func createResource(t *testing.T, f fixture, connectionID int64, identifier string) ResourceView {
	t.Helper()
	connection := f.repo.connections[connectionID]
	actorID := connection.OwnerID
	if connection.OwnerScope == domain.OwnerScopeOrg {
		actorID = 1
	}
	view, err := f.service.CreateResource(context.Background(), actor(actorID), CreateResourceInput{ConnectionID: connectionID, Identifier: slugkit.Slug(identifier), ModelID: "provider/" + identifier, DisplayName: identifier, Modalities: []domain.Modality{domain.ModalityChat}, Capabilities: []domain.Capability{domain.CapabilityTextGeneration}})
	require.NoError(t, err)
	return view
}
