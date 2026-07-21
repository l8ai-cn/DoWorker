package infra

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIResourceRepositoryMetadataWritesPreserveRuntimeRevisions(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	connectionRevision := connection.Revision
	connection.Name = "Renamed"
	connection.CredentialsEncrypted = "rotated"
	connection.Status = airesource.ConnectionStatusUnchecked
	connection.IsEnabled = false
	require.NoError(t, repo.SaveConnectionMetadata(ctx, connection))
	assert.Equal(t, connectionRevision, connection.Revision)

	resource := createAIResource(t, repo, connection.ID, "chat-model", true, airesource.ModalityChat)
	resourceRevision := resource.Revision
	resource.DisplayName = "Renamed model"
	resource.IsEnabled = false
	require.NoError(t, repo.SaveResourceMetadata(ctx, resource))
	assert.Equal(t, resourceRevision, resource.Revision)
}

func TestAIResourceRepositoryMetadataWritesRejectStaleAndRuntimeChanges(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	staleConnection, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	connection.Name = "Current"
	require.NoError(t, repo.SaveConnectionMetadata(ctx, connection))
	staleConnection.Name = "Stale"
	assert.ErrorIs(t, repo.SaveConnectionMetadata(ctx, staleConnection), airesource.ErrConflict)
	connection.BaseURL = "https://runtime-change.example.com"
	assert.Error(t, repo.SaveConnectionMetadata(ctx, connection))

	resource := createAIResource(t, repo, connection.ID, "chat-model", true, airesource.ModalityChat)
	staleResource, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	resource.DisplayName = "Current"
	require.NoError(t, repo.SaveResourceMetadata(ctx, resource))
	staleResource.DisplayName = "Stale"
	assert.ErrorIs(t, repo.SaveResourceMetadata(ctx, staleResource), airesource.ErrConflict)
	resource.ModelID = "runtime-change"
	assert.Error(t, repo.SaveResourceMetadata(ctx, resource))
}

func TestAIResourceRepositoryRuntimeWritesRejectNewerMetadata(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	staleConnection, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	connection.CredentialsEncrypted = "rotated"
	require.NoError(t, repo.SaveConnectionMetadata(ctx, connection))
	staleConnection.BaseURL = "https://runtime-change.example.com"
	assert.ErrorIs(t, repo.SaveConnection(ctx, staleConnection), airesource.ErrConflict)
	resource := createAIResource(t, repo, connection.ID, "chat-model", true, airesource.ModalityChat)
	assert.ErrorIs(
		t,
		repo.DeleteConnection(ctx, staleConnection.ID, staleConnection.Revision, staleConnection.UpdatedAt),
		airesource.ErrConflict,
	)
	storedResource, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	require.NotNil(t, storedResource)

	staleResource, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	resource.IsEnabled = false
	require.NoError(t, repo.SaveResourceMetadata(ctx, resource))
	staleResource.ModelID = "runtime-change"
	assert.ErrorIs(t, repo.SaveResource(ctx, staleResource), airesource.ErrConflict)
	assert.ErrorIs(
		t,
		repo.DeleteResource(ctx, staleResource.ID, staleResource.Revision, staleResource.UpdatedAt),
		airesource.ErrConflict,
	)
}
