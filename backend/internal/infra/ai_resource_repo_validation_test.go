package infra

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIResourceRepositorySetsValidationStateAtomically(t *testing.T) {
	db, repo := setupAIResourceRepository(t)
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	resource := createAIResource(t, repo, connection.ID, "chat-model", true, airesource.ModalityChat)
	at := time.Date(2026, time.July, 10, 8, 0, 0, 0, time.UTC)

	revision, err := repo.SetValidationState(context.Background(), connection.ID, connection.Revision, connection.CredentialsEncrypted, airesource.ConnectionStatusInvalid, at, "credentials rejected")
	require.NoError(t, err)
	assert.Equal(t, connection.Revision, revision)
	storedConnection, err := repo.GetConnectionByID(context.Background(), connection.ID)
	require.NoError(t, err)
	storedResource, err := repo.GetResourceByID(context.Background(), resource.ID)
	require.NoError(t, err)
	assert.Equal(t, connection.Revision, storedConnection.Revision)
	assert.Equal(t, resource.Revision, storedResource.Revision)
	assert.Equal(t, airesource.ConnectionStatusInvalid, storedConnection.Status)
	assert.Equal(t, airesource.ConnectionStatusInvalid, storedResource.Status)
	assert.Equal(t, "credentials rejected", storedConnection.ValidationError)
	assert.Equal(t, "credentials rejected", storedResource.ValidationError)

	require.NoError(t, db.Exec(`CREATE TRIGGER reject_ai_resource_validation BEFORE UPDATE OF status ON model_resources BEGIN SELECT RAISE(FAIL, 'injected'); END`).Error)
	_, err = repo.SetValidationState(context.Background(), connection.ID, revision, connection.CredentialsEncrypted, airesource.ConnectionStatusValid, at.Add(time.Hour), "")
	require.Error(t, err)
	storedConnection, err = repo.GetConnectionByID(context.Background(), connection.ID)
	require.NoError(t, err)
	storedResource, err = repo.GetResourceByID(context.Background(), resource.ID)
	require.NoError(t, err)
	assert.Equal(t, airesource.ConnectionStatusInvalid, storedConnection.Status)
	assert.Equal(t, airesource.ConnectionStatusInvalid, storedResource.Status)
}

func TestAIResourceRepositoryRejectsValidationForRotatedCredentials(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	expectedCredentials := connection.CredentialsEncrypted
	connection.CredentialsEncrypted = "rotated"
	require.NoError(t, repo.SaveConnectionMetadata(context.Background(), connection))

	_, err := repo.SetValidationState(
		context.Background(),
		connection.ID,
		connection.Revision,
		expectedCredentials,
		airesource.ConnectionStatusValid,
		time.Now().UTC(),
		"",
	)
	assert.ErrorIs(t, err, airesource.ErrConflict)
}
