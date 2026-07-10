package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/virtualkey"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestVirtualAPIKeyRepo_GetByIDForScope(t *testing.T) {
	const (
		orgID  = int64(21)
		userID = int64(11)
	)
	db := setupVirtualAPIKeyRepoTestDB(t)
	repo := NewVirtualAPIKeyRepository(db)
	key := &virtualkey.VirtualAPIKey{
		OrganizationID:  orgID,
		UserID:          userID,
		ModelResourceID: 31,
		Name:            "Worker key",
		KeyPrefix:       "dwk_12345678",
		KeyHash:         "key-hash",
		Status:          virtualkey.StatusActive,
	}
	require.NoError(t, db.Create(key).Error)

	t.Run("exact scope", func(t *testing.T) {
		found, err := repo.GetByIDForScope(context.Background(), key.ID, orgID, userID)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, key.ID, found.ID)
	})

	for _, test := range []struct {
		name   string
		id     int64
		orgID  int64
		userID int64
	}{
		{name: "wrong organization", id: key.ID, orgID: 22, userID: userID},
		{name: "wrong user", id: key.ID, orgID: orgID, userID: 12},
		{name: "missing", id: 9999, orgID: orgID, userID: userID},
	} {
		t.Run(test.name, func(t *testing.T) {
			found, err := repo.GetByIDForScope(context.Background(), test.id, test.orgID, test.userID)

			require.NoError(t, err)
			assert.Nil(t, found)
		})
	}
}

func TestVirtualAPIKeyRepo_GetByIDForScope_PropagatesContextError(t *testing.T) {
	db := setupVirtualAPIKeyRepoTestDB(t)
	repo := NewVirtualAPIKeyRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	key, err := repo.GetByIDForScope(ctx, 1, 1, 1)

	assert.Nil(t, key)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestVirtualAPIKeyRepo_UpdateStatusForScope(t *testing.T) {
	db := setupVirtualAPIKeyRepoTestDB(t)
	repo := NewVirtualAPIKeyRepository(db)
	key := &virtualkey.VirtualAPIKey{
		OrganizationID: 21, UserID: 11, ModelResourceID: 31,
		Name: "Worker key", KeyPrefix: "dwk_12345678",
		KeyHash: "key-hash", Status: virtualkey.StatusActive,
	}
	require.NoError(t, db.Create(key).Error)

	updated, err := repo.UpdateStatusForScope(
		context.Background(), key.ID, 21, 12, virtualkey.StatusRevoked,
	)
	require.NoError(t, err)
	assert.False(t, updated)
	require.NoError(t, db.First(key, key.ID).Error)
	assert.Equal(t, virtualkey.StatusActive, key.Status)

	updated, err = repo.UpdateStatusForScope(
		context.Background(), key.ID, 21, 11, virtualkey.StatusRevoked,
	)
	require.NoError(t, err)
	assert.True(t, updated)
	require.NoError(t, db.First(key, key.ID).Error)
	assert.Equal(t, virtualkey.StatusRevoked, key.Status)
}

func setupVirtualAPIKeyRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE virtual_api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			model_resource_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			token_budget INTEGER,
			status TEXT NOT NULL DEFAULT 'active',
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}
