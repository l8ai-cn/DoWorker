package airesource

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConnectionTransitionsStateAndUsesDecryptedCredentials(t *testing.T) {
	f := newFixture()
	view, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "probe-secret"}})
	require.NoError(t, err)
	require.NoError(t, f.service.ValidateConnection(context.Background(), actor(1), view.ID))
	stored := f.repo.connections[view.ID]
	assert.Equal(t, domain.ConnectionStatusValid, stored.Status)
	assert.NotNil(t, stored.LastValidatedAt)
	assert.Empty(t, stored.ValidationError)
	require.Len(t, f.prober.calls, 1)
	assert.Equal(t, "probe-secret", f.prober.calls[0].Credentials["api_key"])
}

func TestValidateConnectionPersistsSafeInvalidState(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	f.prober.err = errors.New("provider rejected credentials: secret-must-not-leak")
	err := f.service.ValidateConnection(context.Background(), actor(1), connection.ID)
	assert.ErrorIs(t, err, ErrValidation)
	stored := f.repo.connections[connection.ID]
	assert.Equal(t, domain.ConnectionStatusInvalid, stored.Status)
	assert.NotNil(t, stored.LastValidatedAt)
	assert.NotContains(t, stored.ValidationError, "secret-must-not-leak")
	assert.NotContains(t, stored.ValidationError, "secret")
}

func TestValidateConnectionPersistsProviderEndpointUnavailable(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	f.prober.err = ErrProviderEndpointUnavailable

	err := f.service.ValidateConnection(context.Background(), actor(1), connection.ID)

	assert.ErrorIs(t, err, ErrProviderEndpointUnavailable)
	assert.Equal(t, domain.ConnectionStatusInvalid, f.repo.connections[connection.ID].Status)
	assert.Equal(t, "provider endpoint unavailable", f.repo.connections[connection.ID].ValidationError)
}

func TestValidateConnectionPersistenceFailureWins(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	f.prober.err = ErrInvalidCredentials
	f.repo.err["SetValidationState"] = errInjected
	err := f.service.ValidateConnection(context.Background(), actor(1), connection.ID)
	assert.ErrorIs(t, err, errInjected)
	assert.NotErrorIs(t, err, ErrValidation)
}

func TestValidateConnectionUpdatesConnectionAndResourcesTogether(t *testing.T) {
	f := newFixture()
	view, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "secret"}})
	require.NoError(t, err)
	resource := createResource(t, f, view.ID, "pending-model")
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.resources[resource.ID].Status)
	require.NoError(t, f.service.ValidateConnection(context.Background(), actor(1), view.ID))
	assert.Equal(t, domain.ConnectionStatusValid, f.repo.connections[view.ID].Status)
	assert.Equal(t, domain.ConnectionStatusValid, f.repo.resources[resource.ID].Status)
}

func TestValidateConnectionAuditFailureIsReturned(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	resource := createResource(t, f, connection.ID, "pending-model")
	f.prober.err = ErrInvalidCredentials
	f.audit.err = errInjected
	f.audit.calls = 0
	f.audit.failOnCall = 2
	err := f.service.ValidateConnection(context.Background(), actor(1), connection.ID)
	assert.ErrorIs(t, err, ErrAudit)
	assert.ErrorIs(t, err, ErrValidation)
	assert.ErrorIs(t, err, errInjected)
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.connections[connection.ID].Status)
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.resources[resource.ID].Status)
}

type blockingProber struct {
	started sync.Once
	ready   chan struct{}
	release chan struct{}
}

func (prober *blockingProber) Probe(context.Context, ProbeInput) error {
	prober.started.Do(func() { close(prober.ready) })
	<-prober.release
	return nil
}

func TestStaleValidationCannotApproveRotatedCredentials(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "old-secret")
	resource := createResource(t, f, connection.ID, "chat-model")
	prober := &blockingProber{ready: make(chan struct{}), release: make(chan struct{})}
	service, err := NewService(Dependencies{Repository: f.repo, Cipher: f.cipher, Members: f.members, Prober: prober, Mutations: f.mutations, Endpoints: allowingEndpoints{}})
	require.NoError(t, err)
	validationDone := make(chan error, 1)
	go func() { validationDone <- service.ValidateConnection(context.Background(), actor(1), connection.ID) }()
	<-prober.ready
	require.NoError(t, service.RotateConnectionCredentials(context.Background(), actor(1), connection.ID, map[string]string{"api_key": "new-secret"}))
	close(prober.release)
	assert.ErrorIs(t, <-validationDone, ErrConflict)
	stored := f.repo.connections[connection.ID]
	assert.Equal(t, domain.ConnectionStatusUnchecked, stored.Status)
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.resources[resource.ID].Status)
	plaintext, decryptErr := f.cipher.Decrypt(stored.CredentialsEncrypted)
	require.NoError(t, decryptErr)
	assert.Contains(t, plaintext, "new-secret")
}

func TestUnsupportedProbeKeepsConnectionAndResourcesUnchecked(t *testing.T) {
	f := newFixture()
	connection, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "fal-main", ProviderKey: "fal", Name: "fal", Credentials: map[string]string{"api_key": "secret"}})
	require.NoError(t, err)
	resource, err := f.service.CreateResource(context.Background(), actor(1), CreateResourceInput{ConnectionID: connection.ID, Identifier: "image-model", ModelID: "fal-ai/image", DisplayName: "Image", Modalities: []domain.Modality{domain.ModalityImage}, Capabilities: []domain.Capability{domain.CapabilityImageGeneration}})
	require.NoError(t, err)
	f.prober.err = ErrProbeUnsupported
	err = f.service.ValidateConnection(context.Background(), actor(1), connection.ID)
	assert.ErrorIs(t, err, ErrProbeUnsupported)
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.connections[connection.ID].Status)
	assert.Equal(t, domain.ConnectionStatusUnchecked, f.repo.resources[resource.ID].Status)
	assert.Equal(t, "validation unavailable", f.repo.connections[connection.ID].ValidationError)
}

func TestAuditEventsContainMetadataAndNeverCredentials(t *testing.T) {
	f := newFixture()
	view, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "audit-secret"}})
	require.NoError(t, err)
	require.NoError(t, f.service.RotateConnectionCredentials(context.Background(), actor(1), view.ID, map[string]string{"api_key": "rotated-secret"}))
	for _, entry := range f.audit.logs {
		encoded, encodeErr := json.Marshal(entry)
		require.NoError(t, encodeErr)
		text := string(encoded)
		assert.NotContains(t, text, "audit-secret")
		assert.NotContains(t, text, "rotated-secret")
		assert.NotContains(t, text, "api_key")
		assert.NotContains(t, text, f.repo.connections[view.ID].CredentialsEncrypted)
		assert.Equal(t, int64(1), *entry.ActorID)
		assert.Equal(t, "request-123", entry.Details["correlation_id"])
		assert.Equal(t, "user", entry.Details["owner_scope"])
		assert.EqualValues(t, 1, entry.Details["owner_id"])
	}
	assert.Equal(t, audit.ActionProviderConnectionCreated, f.audit.logs[0].Action)
	assert.Equal(t, audit.ActionProviderConnectionCredentialsRotated, f.audit.logs[1].Action)
}

func TestCatalogReturnsCopies(t *testing.T) {
	f := newFixture()
	first := f.service.Catalog()
	first[0].DisplayName = "mutated"
	second := f.service.Catalog()
	assert.NotEqual(t, "mutated", second[0].DisplayName)
	_ = time.Second
}
