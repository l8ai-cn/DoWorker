package infra

import (
	"context"
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAIResourceRepositoryEffectiveVisibilityAndDefaults(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()

	userConnection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "user-openai", true)
	orgConnection := createAIConnection(t, repo, airesource.OwnerScopeOrg, 10, "org-openai", true)
	foreignUser := createAIConnection(t, repo, airesource.OwnerScopeUser, 2, "foreign-user", true)
	foreignOrg := createAIConnection(t, repo, airesource.OwnerScopeOrg, 11, "foreign-org", true)
	disabledConnection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "disabled-connection", false)

	userModel := createAIResource(t, repo, userConnection.ID, "user-model", true,
		airesource.ModalityChat, airesource.ModalityImage)
	orgModel := createAIResource(t, repo, orgConnection.ID, "org-model", true, airesource.ModalityChat)
	createAIResource(t, repo, orgConnection.ID, "disabled-model", false, airesource.ModalityChat)
	createAIResource(t, repo, foreignUser.ID, "foreign-user-model", true, airesource.ModalityChat)
	createAIResource(t, repo, foreignOrg.ID, "foreign-org-model", true, airesource.ModalityChat)
	createAIResource(t, repo, disabledConnection.ID, "hidden-model", true, airesource.ModalityChat)

	require.NoError(t, repo.SetDefault(ctx, orgModel.ID, airesource.ModalityChat))
	require.NoError(t, repo.SetDefault(ctx, userModel.ID, airesource.ModalityChat))
	require.NoError(t, repo.SetDefault(ctx, userModel.ID, airesource.ModalityImage))

	chat, err := repo.ListEffective(ctx, 1, 10, []airesource.Modality{airesource.ModalityChat})
	require.NoError(t, err)
	require.Equal(t, []int64{userModel.ID, orgModel.ID}, resourceIDs(chat))
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat}, chat[0].DefaultModalities)
	assert.Empty(t, chat[1].DefaultModalities, "personal default must hide the org default in effective projection")

	image, err := repo.ListEffective(ctx, 1, 10, []airesource.Modality{airesource.ModalityImage})
	require.NoError(t, err)
	require.Equal(t, []int64{userModel.ID}, resourceIDs(image))
	assert.Equal(t, []airesource.Modality{airesource.ModalityImage}, image[0].DefaultModalities)

	all, err := repo.ListEffective(ctx, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, []int64{userModel.ID, orgModel.ID}, resourceIDs(all))

	otherOrg, err := repo.ListEffective(ctx, 1, 11, nil)
	require.NoError(t, err)
	assert.NotContains(t, resourceIDs(otherOrg), orgModel.ID)
	otherUser, err := repo.ListEffective(ctx, 2, 10, nil)
	require.NoError(t, err)
	assert.NotContains(t, resourceIDs(otherUser), userModel.ID)

	_, err = repo.ListEffective(ctx, 0, 10, nil)
	require.Error(t, err, "organization resources require an authenticated user")
	_, err = repo.ListEffective(ctx, 1, -1, nil)
	require.Error(t, err, "negative organization IDs are invalid")
}

func TestAIResourceRepositoryScopedUniquenessAndCRUD(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()

	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	duplicate := validAIConnection(airesource.OwnerScopeUser, 1, "openai-main", true)
	require.Error(t, repo.CreateConnection(ctx, duplicate))
	otherOwner := validAIConnection(airesource.OwnerScopeUser, 2, "openai-main", true)
	require.NoError(t, repo.CreateConnection(ctx, otherOwner))

	loaded, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	require.Equal(t, "encrypted", loaded.CredentialsEncrypted)
	loaded.Name = "Renamed connection"
	originalCreatedAt := loaded.CreatedAt
	loaded.CreatedAt = loaded.CreatedAt.AddDate(-1, 0, 0)
	loaded.CreatedBy = 2
	require.NoError(t, repo.SaveConnection(ctx, loaded))
	owned, err := repo.ListConnectionsByOwner(ctx, airesource.OwnerScopeUser, 1)
	require.NoError(t, err)
	require.Len(t, owned, 1)
	assert.Equal(t, "Renamed connection", owned[0].Name)
	assert.Equal(t, int64(1), owned[0].CreatedBy)
	assert.True(t, originalCreatedAt.Equal(owned[0].CreatedAt))

	missingConnection := *loaded
	missingConnection.ID = 999
	missingConnection.Identifier = slugkit.Slug("missing-connection")
	require.ErrorIs(t, repo.SaveConnection(ctx, &missingConnection), gorm.ErrRecordNotFound)
	missingLoaded, err := repo.GetConnectionByID(ctx, missingConnection.ID)
	require.NoError(t, err)
	assert.Nil(t, missingLoaded, "SaveConnection must not resurrect a deleted ID")

	resource := createAIResource(t, repo, connection.ID, "gpt-main", true, airesource.ModalityChat)
	duplicateResource := validAIResource(connection.ID, "gpt-main", true, airesource.ModalityChat)
	require.Error(t, repo.CreateResource(ctx, duplicateResource))
	otherConnectionResource := validAIResource(otherOwner.ID, "gpt-main", true, airesource.ModalityChat)
	require.NoError(t, repo.CreateResource(ctx, otherConnectionResource))

	resource.DisplayName = "GPT renamed"
	resourceCreatedAt := resource.CreatedAt
	resource.CreatedAt = resource.CreatedAt.AddDate(-1, 0, 0)
	require.NoError(t, repo.SaveResource(ctx, resource))
	loadedResource, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	assert.Equal(t, "GPT renamed", loadedResource.DisplayName)
	assert.True(t, resourceCreatedAt.Equal(loadedResource.CreatedAt))
	listed, err := repo.ListResourcesByConnection(ctx, connection.ID)
	require.NoError(t, err)
	require.Equal(t, []int64{resource.ID}, resourceIDs(listed))

	missingResource := *resource
	missingResource.ID = 999
	missingResource.Identifier = slugkit.Slug("missing-resource")
	require.ErrorIs(t, repo.SaveResource(ctx, &missingResource), gorm.ErrRecordNotFound)
	missingResourceLoaded, err := repo.GetResourceByID(ctx, missingResource.ID)
	require.NoError(t, err)
	assert.Nil(t, missingResourceLoaded, "SaveResource must not resurrect a deleted ID")
}

func TestAIResourceRepositoryPerModalityDefaultsAndParentUpdates(t *testing.T) {
	db, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	multimodal := createAIResource(t, repo, connection.ID, "multi", true,
		airesource.ModalityChat, airesource.ModalityImage)
	chatOnly := createAIResource(t, repo, connection.ID, "chat-only", true, airesource.ModalityChat)

	require.NoError(t, repo.SetDefault(ctx, multimodal.ID, airesource.ModalityChat))
	require.NoError(t, repo.SetDefault(ctx, multimodal.ID, airesource.ModalityImage))
	require.NoError(t, repo.SetDefault(ctx, chatOnly.ID, airesource.ModalityChat))
	chatOnly.DefaultModalities = []airesource.Modality{airesource.ModalityImage}
	chatOnly.DisplayName = "Chat renamed"
	require.NoError(t, repo.SaveResource(ctx, chatOnly),
		"caller projection must not block ordinary persisted-field updates")
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat}, chatOnly.DefaultModalities)

	noDefault := createAIResource(t, repo, connection.ID, "no-default", true, airesource.ModalityChat)
	noDefault.DefaultModalities = []airesource.Modality{airesource.ModalityChat}
	require.NoError(t, repo.SaveResource(ctx, noDefault))
	assert.Empty(t, noDefault.DefaultModalities, "SaveResource must reload DB-authoritative defaults")

	resources, err := repo.ListResourcesByConnection(ctx, connection.ID)
	require.NoError(t, err)
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat}, findAIResource(t, resources, chatOnly.ID).DefaultModalities)
	assert.Equal(t, []airesource.Modality{airesource.ModalityImage}, findAIResource(t, resources, multimodal.ID).DefaultModalities)
	require.Error(t, repo.SetDefault(ctx, chatOnly.ID, airesource.ModalityAudio))

	multimodal.Modalities = []airesource.Modality{airesource.ModalityChat}
	multimodal.DefaultModalities = []airesource.Modality{airesource.ModalityAudio}
	require.Error(t, repo.SaveResource(ctx, multimodal), "removing a defaulted modality must be rejected")
	connection.OwnerID = 2
	require.Error(t, repo.SaveConnection(ctx, connection), "connection ownership must be immutable")
	require.Error(t, db.Exec("DELETE FROM users WHERE id = ?", 1).Error,
		"deleting an owner must not orphan its provider connections")
	createAIConnection(t, repo, airesource.OwnerScopeOrg, 10, "org-owner", true)
	require.Error(t, db.Exec("DELETE FROM organizations WHERE id = ?", 10).Error,
		"deleting an org must not orphan its provider connections")

	result := db.Exec(`INSERT INTO model_resource_defaults
		(owner_scope, owner_id, modality, model_resource_id) VALUES (?, ?, ?, ?)`,
		airesource.OwnerScopeOrg, 10, airesource.ModalityChat, chatOnly.ID)
	require.Error(t, result.Error, "database constraint must reject defaults owned by another scope")
}

func TestAIResourceRepositoryDeleteCleanupValidationAndQueryErrors(t *testing.T) {
	db, repo := setupAIResourceRepository(t)
	ctx := context.Background()

	invalidConnection := validAIConnection(airesource.OwnerScopeUser, 1, "bad_name", true)
	require.Error(t, repo.CreateConnection(ctx, invalidConnection))
	invalidProvider := validAIConnection(airesource.OwnerScopeUser, 1, "valid-name", true)
	invalidProvider.ProviderKey = airesource.ProviderDefinition{}.Key
	require.Error(t, repo.CreateConnection(ctx, invalidProvider))
	orphanOwner := validAIConnection(airesource.OwnerScopeUser, 999, "orphan-owner", true)
	require.Error(t, repo.CreateConnection(ctx, orphanOwner))

	connection := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "openai-main", true)
	invalidResource := validAIResource(connection.ID, "bad_name", true, airesource.ModalityChat)
	require.Error(t, repo.CreateResource(ctx, invalidResource))
	emptyModelID := validAIResource(connection.ID, "empty-model", true, airesource.ModalityChat)
	emptyModelID.ModelID = " "
	require.Error(t, repo.CreateResource(ctx, emptyModelID))

	resource := createAIResource(t, repo, connection.ID, "gpt-main", true,
		airesource.ModalityChat, airesource.ModalityImage)
	require.NoError(t, repo.SetDefault(ctx, resource.ID, airesource.ModalityChat))
	require.NoError(t, repo.SetDefault(ctx, resource.ID, airesource.ModalityImage))
	require.NoError(t, repo.DeleteResource(ctx, resource.ID, resource.Revision, resource.UpdatedAt))
	assertTableCount(t, db, "model_resource_defaults", 0)
	loaded, err := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded)

	replacement := createAIResource(t, repo, connection.ID, "replacement", true, airesource.ModalityChat)
	require.NoError(t, repo.SetDefault(ctx, replacement.ID, airesource.ModalityChat))
	require.NoError(t, repo.DeleteConnection(ctx, connection.ID, connection.Revision, connection.UpdatedAt))
	assertTableCount(t, db, "model_resource_defaults", 0)
	assertTableCount(t, db, "model_resources", 0)

	require.NoError(t, db.Exec("DROP TABLE model_resources").Error)
	_, err = repo.ListEffective(ctx, 1, 10, nil)
	require.Error(t, err, "query errors must not degrade to an empty resource list")
}

func TestAIResourceJSONListBindsAsText(t *testing.T) {
	value, err := jsonStringList{"chat", "image"}.Value()
	require.NoError(t, err)
	assert.Equal(t, `["chat","image"]`, value)
}

func TestAIResourceRepositorySaveRejectsConnectionReassignment(t *testing.T) {
	_, repo := setupAIResourceRepository(t)
	ctx := context.Background()
	first := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "first-connection", true)
	second := createAIConnection(t, repo, airesource.OwnerScopeUser, 1, "second-connection", true)
	resource := createAIResource(t, repo, first.ID, "fixed-resource", true, airesource.ModalityChat)
	resource.ProviderConnectionID = second.ID

	err := repo.SaveResource(ctx, resource)
	require.Error(t, err)
	stored, loadErr := repo.GetResourceByID(ctx, resource.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, first.ID, stored.ProviderConnectionID)
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func setupAIResourceRepository(t *testing.T) (*gorm.DB, airesource.Repository) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	for _, statement := range []string{
		`INSERT INTO users(id, email, username) VALUES (1, 'one@example.com', 'user-one')`,
		`INSERT INTO users(id, email, username) VALUES (2, 'two@example.com', 'user-two')`,
		`INSERT INTO organizations(id, name, slug) VALUES (10, 'Org Ten', 'org-ten')`,
		`INSERT INTO organizations(id, name, slug) VALUES (11, 'Org Eleven', 'org-eleven')`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db, NewAIResourceRepository(db)
}

func validAIConnection(scope airesource.OwnerScope, ownerID int64, identifier string, enabled bool) *airesource.Connection {
	return &airesource.Connection{
		OwnerScope: scope, OwnerID: ownerID, Identifier: slugkit.Slug(identifier),
		ProviderKey: slugkit.Slug("openai"), Name: identifier, BaseURL: "https://api.openai.com",
		CredentialsEncrypted: "encrypted", ConfiguredFields: []string{"api-key"},
		Status: airesource.ConnectionStatusValid, IsEnabled: enabled, CreatedBy: 1,
	}
}

func createAIConnection(t *testing.T, repo airesource.Repository, scope airesource.OwnerScope, ownerID int64, identifier string, enabled bool) *airesource.Connection {
	t.Helper()
	connection := validAIConnection(scope, ownerID, identifier, enabled)
	require.NoError(t, repo.CreateConnection(context.Background(), connection))
	return connection
}

func validAIResource(connectionID int64, identifier string, enabled bool, modalities ...airesource.Modality) *airesource.ModelResource {
	return &airesource.ModelResource{
		ProviderConnectionID: connectionID, Identifier: slugkit.Slug(identifier), ModelID: identifier,
		DisplayName: identifier, Modalities: modalities, Capabilities: []airesource.Capability{airesource.CapabilityTextGeneration},
		Status: airesource.ConnectionStatusValid, IsEnabled: enabled,
	}
}

func createAIResource(t *testing.T, repo airesource.Repository, connectionID int64, identifier string, enabled bool, modalities ...airesource.Modality) *airesource.ModelResource {
	t.Helper()
	resource := validAIResource(connectionID, identifier, enabled, modalities...)
	require.NoError(t, repo.CreateResource(context.Background(), resource))
	return resource
}

func resourceIDs(resources []*airesource.ModelResource) []int64 {
	ids := make([]int64, len(resources))
	for index, resource := range resources {
		ids[index] = resource.ID
	}
	return ids
}

func findAIResource(t *testing.T, resources []*airesource.ModelResource, id int64) *airesource.ModelResource {
	t.Helper()
	for _, resource := range resources {
		if resource.ID == id {
			return resource
		}
	}
	t.Fatalf("resource %d not found", id)
	return nil
}

func assertTableCount(t *testing.T, db *gorm.DB, table string, expected int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Table(table).Count(&count).Error)
	assert.Equal(t, expected, count)
}
