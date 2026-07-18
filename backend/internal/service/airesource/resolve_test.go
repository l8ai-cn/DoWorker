package airesource

import (
	"context"
	"errors"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveExactReturnsOnlySubmittedResourceAndCredentials(t *testing.T) {
	f := newFixture()
	connectionA := createValidConnection(t, f, domain.OwnerScopeUser, 1, "connection-a", "secret-a")
	resourceA := createResource(t, f, connectionA.ID, "model-a")
	connectionB := createValidConnection(t, f, domain.OwnerScopeUser, 1, "connection-b", "secret-b")
	resourceB := createResource(t, f, connectionB.ID, "model-b")
	require.NoError(t, f.service.SetDefault(context.Background(), actor(1), resourceA.ID, domain.ModalityChat))

	resolved, err := f.service.ResolveExact(context.Background(), actor(1), 0, resourceB.ID, chatRequirements())
	require.NoError(t, err)
	assert.Equal(t, resourceB.ID, resolved.Resource.ID)
	assert.Equal(t, connectionB.ID, resolved.Connection.ID)
	assert.Equal(t, map[string]string{"api_key": "secret-b"}, resolved.Credentials)
	assert.NotEqual(t, resourceA.ID, resolved.Resource.ID)
}

func TestResolveMetadataDoesNotDecryptCredentials(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "connection-a", "secret-a")
	resource := createResource(t, f, connection.ID, "model-a")
	service, err := NewService(Dependencies{
		Repository: f.repo,
		Cipher: failingCipher{
			decryptErr: errors.New("metadata resolution must not decrypt credentials"),
		},
		Members: f.members, Prober: f.prober, Mutations: f.mutations,
		Endpoints: allowingEndpoints{},
	})
	require.NoError(t, err)

	resolved, err := service.ResolveMetadata(
		context.Background(), actor(1), 0, resource.ID, chatRequirements(),
	)

	require.NoError(t, err)
	assert.Equal(t, resource.ID, resolved.Resource.ID)
	assert.Empty(t, resolved.Credentials)
}

func TestResolveExactPreservesRuntimeRevisionAcrossCredentialRotation(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "connection-a", "old-secret")
	resource := createResource(t, f, connection.ID, "model-a")
	connectionRevision := f.repo.connections[connection.ID].Revision
	resourceRevision := f.repo.resources[resource.ID].Revision

	require.NoError(t, f.service.RotateConnectionCredentials(
		context.Background(),
		actor(1),
		connection.ID,
		map[string]string{"api_key": "new-secret"},
	))
	require.NoError(t, f.service.ValidateConnection(context.Background(), actor(1), connection.ID))
	resolved, err := f.service.ResolveExact(
		context.Background(),
		actor(1),
		0,
		resource.ID,
		chatRequirements(),
	)

	require.NoError(t, err)
	assert.Equal(t, connectionRevision, resolved.Connection.Revision)
	assert.Equal(t, resourceRevision, resolved.Resource.Revision)
	assert.Equal(t, "new-secret", resolved.Credentials["api_key"])
}

func TestResolveExactTreatsEnableStateAsOperationalMetadata(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "connection-a", "secret")
	resource := createResource(t, f, connection.ID, "model-a")
	connectionRevision := f.repo.connections[connection.ID].Revision
	resourceRevision := f.repo.resources[resource.ID].Revision

	require.NoError(t, f.service.SetConnectionEnabled(context.Background(), actor(1), connection.ID, false))
	_, err := f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrDisabled)
	require.NoError(t, f.service.SetConnectionEnabled(context.Background(), actor(1), connection.ID, true))
	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())

	require.NoError(t, err)
	assert.Equal(t, connectionRevision, f.repo.connections[connection.ID].Revision)
	assert.Equal(t, resourceRevision, f.repo.resources[resource.ID].Revision)
}

func TestResolveExactRejectsVisibilityAndInvalidStates(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*fixture, int64, int64)
		expected error
	}{
		{"connection disabled", func(f *fixture, connectionID, _ int64) { f.repo.connections[connectionID].IsEnabled = false }, ErrDisabled},
		{"resource disabled", func(f *fixture, _, resourceID int64) { f.repo.resources[resourceID].IsEnabled = false }, ErrDisabled},
		{"connection unchecked", func(f *fixture, connectionID, _ int64) {
			f.repo.connections[connectionID].Status = domain.ConnectionStatusUnchecked
		}, ErrUnchecked},
		{"resource unchecked", func(f *fixture, _, resourceID int64) {
			f.repo.resources[resourceID].Status = domain.ConnectionStatusUnchecked
		}, ErrUnchecked},
		{"connection invalid", func(f *fixture, connectionID, _ int64) {
			f.repo.connections[connectionID].Status = domain.ConnectionStatusInvalid
		}, ErrUnhealthy},
		{"resource invalid", func(f *fixture, _, resourceID int64) {
			f.repo.resources[resourceID].Status = domain.ConnectionStatusInvalid
		}, ErrUnhealthy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newFixture()
			connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
			resource := createResource(t, f, connection.ID, "model-b")
			test.mutate(&f, connection.ID, resource.ID)
			_, err := f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())
			assert.ErrorIs(t, err, test.expected)
		})
	}

	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "model-b")
	_, err := f.service.ResolveExact(context.Background(), actor(2), 0, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrForbidden)
	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, ResolutionRequirements{Modality: domain.ModalityVideo, Capability: domain.CapabilityVideoGeneration, AllowedProtocolAdapters: []string{"openai-compatible"}})
	assert.ErrorIs(t, err, ErrIncompatibleModality)
}

func TestResolveExactRequiresOrganizationMembership(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeOrg, 10, "org-main", "secret")
	resource := createResource(t, f, connection.ID, "org-model")
	_, err := f.service.ResolveExact(context.Background(), actor(99), 10, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrForbidden)
	_, err = f.service.ResolveExact(context.Background(), actor(3), 11, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrForbidden)
	resolved, err := f.service.ResolveExact(context.Background(), actor(3), 10, resource.ID, chatRequirements())
	require.NoError(t, err)
	assert.Equal(t, resource.ID, resolved.Resource.ID)
}

func TestResolveExactValidatesWorkerOrganizationForPersonalResource(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "user-main", "secret")
	resource := createResource(t, f, connection.ID, "user-model")
	_, err := f.service.ResolveExact(context.Background(), actor(1), 11, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrForbidden)
	f.members.err = errInjected
	_, err = f.service.ResolveExact(context.Background(), actor(1), 10, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, errInjected)
}

func TestResolveExactReturnsTypedDecryptAndNotFoundErrors(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "model-b")
	f.repo.connections[connection.ID].CredentialsEncrypted = "broken"
	_, err := f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())
	assert.ErrorIs(t, err, ErrDecrypt)
	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, 999, chatRequirements())
	assert.ErrorIs(t, err, ErrNotFound)
	f.repo.err["GetResourceByID"] = errInjected
	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, chatRequirements())
	assert.True(t, errors.Is(err, errInjected))
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestResolveExactRejectsCapabilityAndProtocolAdapterMismatch(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "model-b")
	multimodal, err := f.service.CreateResource(context.Background(), actor(1), CreateResourceInput{ConnectionID: connection.ID, Identifier: "multimodal-model", ModelID: "gpt-multimodal", DisplayName: "Multimodal", Modalities: []domain.Modality{domain.ModalityMultimodal}, Capabilities: []domain.Capability{domain.CapabilityTextGeneration}})
	require.NoError(t, err)

	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, multimodal.ID, ResolutionRequirements{Modality: domain.ModalityMultimodal, Capability: domain.CapabilityVisionInput, AllowedProtocolAdapters: []string{"openai-compatible"}})
	assert.ErrorIs(t, err, ErrIncompatibleCapability)
	_, err = f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, ResolutionRequirements{Modality: domain.ModalityChat, Capability: domain.CapabilityTextGeneration, AllowedProtocolAdapters: []string{"anthropic"}})
	assert.ErrorIs(t, err, ErrIncompatibleProtocolAdapter)
}

func TestResolveExactRejectsIncompleteRequirements(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "model-b")
	for _, requirements := range []ResolutionRequirements{
		{Capability: domain.CapabilityTextGeneration, AllowedProtocolAdapters: []string{"openai-compatible"}},
		{Modality: domain.ModalityChat, AllowedProtocolAdapters: []string{"openai-compatible"}},
		{Modality: domain.ModalityChat, Capability: domain.CapabilityTextGeneration},
		{Modality: domain.ModalityChat, Capability: domain.CapabilityTextGeneration, AllowedProtocolAdapters: []string{"OpenAI_Compatible"}},
	} {
		_, err := f.service.ResolveExact(context.Background(), actor(1), 0, resource.ID, requirements)
		assert.ErrorIs(t, err, ErrInvalidRequirements)
	}
}

func chatRequirements() ResolutionRequirements {
	return ResolutionRequirements{Modality: domain.ModalityChat, Capability: domain.CapabilityTextGeneration, AllowedProtocolAdapters: []string{"openai-compatible"}}
}
