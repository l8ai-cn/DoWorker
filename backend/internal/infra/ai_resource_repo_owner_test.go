package infra

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIResourceRepositoryListsAllResourcesForOneOwner(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	first := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "first-main", true)
	second := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "second-main", false)
	foreign := createAIConnection(t, repo, airesource.OwnerScopeUser, 2, "foreign-main", true)
	firstResource := createAIResource(t, repo, first.ID, "first-model", true, airesource.ModalityChat)
	secondResource := createAIResource(t, repo, second.ID, "second-model", false, airesource.ModalityChat)
	createAIResource(t, repo, foreign.ID, "foreign-model", true, airesource.ModalityChat)

	resources, err := repo.ListResourcesByOwner(context.Background(), airesource.OwnerScopeUser, 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{firstResource.ID, secondResource.ID}, resourceIDs(resources))
}

func TestAIResourceRepositoryReturnsTypedScopedIdentifierConflicts(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	first := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "first-main", true)
	duplicate := validAIConnection(airesource.OwnerScopeUser, 1, "first-main", true)
	assert.ErrorIs(t, repo.CreateConnection(context.Background(), duplicate), airesource.ErrConflict)

	second := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "second-main", true)
	second.Identifier = first.Identifier
	assert.ErrorIs(t, repo.SaveConnection(context.Background(), second), airesource.ErrConflict)

	firstResource := createAIResource(t, repo, first.ID, "first-model", true, airesource.ModalityChat)
	duplicateResource := validAIResource(first.ID, "first-model", true, airesource.ModalityChat)
	assert.ErrorIs(t, repo.CreateResource(context.Background(), duplicateResource), airesource.ErrConflict)

	secondResource := createAIResource(t, repo, first.ID, "second-model", true, airesource.ModalityChat)
	secondResource.Identifier = firstResource.Identifier
	assert.ErrorIs(t, repo.SaveResource(context.Background(), secondResource), airesource.ErrConflict)
}

func TestAIResourceRepositoryRejectsStaleRevisions(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	first, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	stale, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	first.Name = "Renamed"
	require.NoError(t, repo.SaveConnection(ctx, first))
	stale.IsEnabled = false
	assert.ErrorIs(t, repo.SaveConnection(ctx, stale), airesource.ErrConflict)
	_, err = repo.SetValidationState(ctx, connection.ID, stale.Revision, stale.CredentialsEncrypted, airesource.ConnectionStatusValid, time.Now(), "")
	assert.ErrorIs(t, err, airesource.ErrConflict)

	resource := createAIResource(t, repo, connection.ID, "chat-model", true, airesource.ModalityChat)
	resourceFirst, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	resourceStale, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	resourceFirst.DisplayName = "Renamed"
	require.NoError(t, repo.SaveResource(ctx, resourceFirst))
	resourceStale.IsEnabled = false
	assert.ErrorIs(t, repo.SaveResource(ctx, resourceStale), airesource.ErrConflict)
}

func TestInvalidPersonalDefaultDoesNotHideValidOrganizationDefault(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	userConnection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "user-main", true)
	orgConnection := createAIConnection(t, repo, airesource.OwnerScopeOrg, 10, "org-main", true)
	userResource := createAIResource(t, repo, userConnection.ID, "user-model", true, airesource.ModalityChat)
	orgResource := createAIResource(t, repo, orgConnection.ID, "org-model", true, airesource.ModalityChat)
	require.NoError(t, repo.SetDefault(ctx, userResource.ID, airesource.ModalityChat))
	require.NoError(t, repo.SetDefault(ctx, orgResource.ID, airesource.ModalityChat))
	userResource.Status = airesource.ConnectionStatusInvalid
	require.NoError(t, repo.SaveResource(ctx, userResource))

	resources, err := repo.ListEffective(ctx, 1, 10, []airesource.Modality{airesource.ModalityChat})
	require.NoError(t, err)
	require.Equal(t, []int64{orgResource.ID}, resourceIDs(resources))
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat}, resources[0].DefaultModalities)
}
