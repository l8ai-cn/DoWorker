package airesource

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceRejectsMissingDependencies(t *testing.T) {
	f := newFixture()
	valid := Dependencies{Repository: f.repo, Cipher: f.cipher, Members: f.members, Prober: f.prober, Mutations: f.mutations, Endpoints: allowingEndpoints{}}
	tests := []struct {
		name   string
		mutate func(*Dependencies)
	}{
		{"repository", func(d *Dependencies) { d.Repository = nil }},
		{"cipher", func(d *Dependencies) { d.Cipher = nil }},
		{"members", func(d *Dependencies) { d.Members = nil }},
		{"prober", func(d *Dependencies) { d.Prober = nil }},
		{"mutation runner", func(d *Dependencies) { d.Mutations = nil }},
		{"endpoint policy", func(d *Dependencies) { d.Endpoints = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deps := valid
			test.mutate(&deps)
			service, err := NewService(deps)
			assert.Nil(t, service)
			assert.Error(t, err)
		})
	}
}

func TestCreateConnectionEncryptsWriteOnlyCredentials(t *testing.T) {
	f := newFixture()
	created, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{
		OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: slugkit.Slug("openai-main"),
		ProviderKey: slugkit.Slug("openai"), Name: "OpenAI main", Credentials: map[string]string{"api_key": "top-secret-token"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"api_key"}, created.ConfiguredFields)
	encoded, err := json.Marshal(created)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "top-secret-token")
	stored := f.repo.connections[created.ID]
	assert.NotContains(t, stored.CredentialsEncrypted, "top-secret-token")
	plaintext, err := f.cipher.Decrypt(stored.CredentialsEncrypted)
	require.NoError(t, err)
	assert.JSONEq(t, `{"api_key":"top-secret-token"}`, plaintext)
	assert.Equal(t, domain.ConnectionStatusUnchecked, stored.Status)
}

func TestConnectionCredentialValidation(t *testing.T) {
	f := newFixture()
	base := CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: slugkit.Slug("openai-main"), ProviderKey: slugkit.Slug("openai"), Name: "OpenAI"}
	tests := []struct {
		name        string
		credentials map[string]string
	}{
		{"missing required", map[string]string{}},
		{"empty required", map[string]string{"api_key": "  "}},
		{"unknown field", map[string]string{"api_key": "secret", "organization": "leak"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := base
			input.Identifier = slugkit.Slug("openai-" + strings.ReplaceAll(test.name, " ", "-"))
			input.Credentials = test.credentials
			_, err := f.service.CreateConnection(context.Background(), actor(1), input)
			assert.ErrorIs(t, err, ErrInvalidCredentials)
		})
	}
}

func TestRotateCredentialsRequiresRevalidation(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "old-secret")
	revision := f.repo.connections[connection.ID].Revision
	require.NoError(t, f.service.RotateConnectionCredentials(context.Background(), actor(1), connection.ID, map[string]string{"api_key": "new-secret"}))
	stored := f.repo.connections[connection.ID]
	assert.Equal(t, revision, stored.Revision)
	assert.Equal(t, domain.ConnectionStatusUnchecked, stored.Status)
	assert.Nil(t, stored.LastValidatedAt)
	assert.Empty(t, stored.ValidationError)
	_, err := f.service.ResolveExact(context.Background(), actor(1), 0, createResource(t, f, connection.ID, "model-b").ID, chatRequirements())
	assert.ErrorIs(t, err, ErrUnchecked)
}

func TestOwnerAndOrganizationPermissions(t *testing.T) {
	f := newFixture()
	userConnection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "user-main", "secret")
	_, err := f.service.ListOwnerConnections(context.Background(), actor(2), domain.OwnerScopeUser, 1)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.ErrorIs(t, f.service.SetConnectionEnabled(context.Background(), actor(2), userConnection.ID, false), ErrForbidden)

	orgConnection := createValidConnection(t, f, domain.OwnerScopeOrg, 10, "org-main", "secret")
	memberViews, err := f.service.ListOwnerConnections(context.Background(), actor(3), domain.OwnerScopeOrg, 10)
	require.NoError(t, err)
	require.Len(t, memberViews, 1)
	assert.False(t, memberViews[0].CanManage)
	assert.ErrorIs(t, f.service.SetConnectionEnabled(context.Background(), actor(3), orgConnection.ID, false), ErrForbidden)
	require.NoError(t, f.service.SetConnectionEnabled(context.Background(), actor(2), orgConnection.ID, false))
	_, err = f.service.ListOwnerConnections(context.Background(), actor(99), domain.OwnerScopeOrg, 10)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestMembershipLookupFailureIsNotDisguised(t *testing.T) {
	f := newFixture()
	f.members.err = errInjected
	_, err := f.service.ListOwnerConnections(context.Background(), actor(1), domain.OwnerScopeOrg, 10)
	assert.ErrorIs(t, err, errInjected)
	assert.NotErrorIs(t, err, ErrForbidden)
}

func TestActorAndOwnerIDsMustBePositive(t *testing.T) {
	f := newFixture()
	_, err := f.service.ListEffective(context.Background(), actor(0), 0, nil)
	assert.ErrorIs(t, err, ErrInvalidOwner)
	_, err = f.service.ListOwnerConnections(context.Background(), actor(1), domain.OwnerScopeOrg, 0)
	assert.ErrorIs(t, err, ErrInvalidOwner)
	_, err = f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: -1})
	assert.ErrorIs(t, err, ErrInvalidOwner)
}

func TestUpdateConnectionCredentialsAreWriteOnlyAndResetStatus(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "old-secret")
	revision := f.repo.connections[connection.ID].Revision
	view, err := f.service.UpdateConnection(context.Background(), actor(1), connection.ID, UpdateConnectionInput{Name: "Renamed", Credentials: map[string]string{"api_key": "replacement"}})
	require.NoError(t, err)
	assert.Equal(t, revision, f.repo.connections[connection.ID].Revision)
	assert.Equal(t, "Renamed", view.Name)
	assert.Equal(t, domain.ConnectionStatusUnchecked, view.Status)
	encoded, _ := json.Marshal(view)
	assert.NotContains(t, string(encoded), "replacement")
	plaintext, err := f.cipher.Decrypt(f.repo.connections[connection.ID].CredentialsEncrypted)
	require.NoError(t, err)
	assert.Contains(t, plaintext, "replacement")
}

func TestConnectionRevisionOnlyChangesForRuntimeConfiguration(t *testing.T) {
	f := newFixture()
	connection, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{
		OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "custom-main",
		ProviderKey: "custom-openai-compatible", Name: "Custom",
		BaseURL: "https://first.example.com/v1", Credentials: map[string]string{"api_key": "secret"},
	})
	require.NoError(t, err)
	revision := f.repo.connections[connection.ID].Revision
	require.NoError(t, f.service.SetConnectionEnabled(context.Background(), actor(1), connection.ID, false))
	assert.Equal(t, revision, f.repo.connections[connection.ID].Revision)

	_, err = f.service.UpdateConnection(
		context.Background(),
		actor(1),
		connection.ID,
		UpdateConnectionInput{BaseURL: "https://second.example.com/v1"},
	)
	require.NoError(t, err)
	assert.Equal(t, revision+1, f.repo.connections[connection.ID].Revision)
}

func TestConnectionDatabaseAndAuditErrorsAreNotSwallowed(t *testing.T) {
	f := newFixture()
	f.repo.err["CreateConnection"] = errInjected
	_, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "secret"}})
	assert.ErrorIs(t, err, errInjected)

	f = newFixture()
	f.audit.err = errInjected
	_, err = f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", Credentials: map[string]string{"api_key": "secret"}})
	assert.ErrorIs(t, err, ErrAudit)
	assert.ErrorIs(t, err, errInjected)
	assert.Empty(t, f.repo.connections)
}

func TestConnectionAuditFailureRollsBackUpdateAndDelete(t *testing.T) {
	f := newFixture()
	connection := createValidConnection(t, f, domain.OwnerScopeUser, 1, "openai-main", "secret")
	f.audit.err = errInjected
	_, err := f.service.UpdateConnection(context.Background(), actor(1), connection.ID, UpdateConnectionInput{Name: "Must rollback"})
	assert.ErrorIs(t, err, ErrAudit)
	assert.Equal(t, "openai-main", f.repo.connections[connection.ID].Name)
	err = f.service.DeleteConnection(context.Background(), actor(1), connection.ID)
	assert.ErrorIs(t, err, ErrAudit)
	assert.NotNil(t, f.repo.connections[connection.ID])
}

func createValidConnection(t *testing.T, f fixture, scope domain.OwnerScope, ownerID int64, identifier, secret string) ConnectionView {
	t.Helper()
	actorID := ownerID
	if scope == domain.OwnerScopeOrg {
		actorID = 1
	}
	view, err := f.service.CreateConnection(context.Background(), actor(actorID), CreateConnectionInput{OwnerScope: scope, OwnerID: ownerID, Identifier: slugkit.Slug(identifier), ProviderKey: slugkit.Slug("openai"), Name: identifier, Credentials: map[string]string{"api_key": secret}})
	require.NoError(t, err)
	stored := f.repo.connections[view.ID]
	stored.Status = domain.ConnectionStatusValid
	f.repo.connections[view.ID] = stored
	return view
}

func TestOrganizationReaderContractUsesDomainMember(t *testing.T) {
	var reader OrganizationMemberReader = &memberReader{}
	_, err := reader.GetMember(context.Background(), 1, 1)
	assert.True(t, errors.Is(err, organization.ErrMemberNotFound))
	_ = crypto.ErrDecryptionFailed
}
