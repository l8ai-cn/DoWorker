package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAIModelRepo_GetVisibleByID(t *testing.T) {
	const (
		userID      = int64(11)
		orgID       = int64(21)
		otherUserID = int64(12)
		otherOrgID  = int64(22)
	)
	db := setupAIModelRepoTestDB(t)
	repo := NewAIModelRepository(db)

	models := map[string]*aimodel.AIModel{
		"same-org shared":      newAIModel(pointerTo(orgID), nil),
		"current-user private": newAIModel(nil, pointerTo(userID)),
		"other-org":            newAIModel(pointerTo(otherOrgID), nil),
		"other-user":           newAIModel(nil, pointerTo(otherUserID)),
		"disabled":             newAIModel(pointerTo(orgID), nil),
	}
	for _, model := range models {
		require.NoError(t, db.Create(model).Error)
	}
	require.NoError(t, db.Model(models["disabled"]).Update("is_enabled", false).Error)

	for _, name := range []string{"same-org shared", "current-user private"} {
		t.Run(name, func(t *testing.T) {
			model := models[name]
			visible, err := repo.GetVisibleByID(context.Background(), model.ID, userID, orgID)
			require.NoError(t, err)
			require.NotNil(t, visible)
			assert.Equal(t, model.ID, visible.ID)
		})
	}

	for _, name := range []string{"other-org", "other-user", "disabled"} {
		t.Run(name, func(t *testing.T) {
			visible, err := repo.GetVisibleByID(context.Background(), models[name].ID, userID, orgID)
			require.NoError(t, err)
			assert.Nil(t, visible)
		})
	}

	t.Run("missing", func(t *testing.T) {
		visible, err := repo.GetVisibleByID(context.Background(), 9999, userID, orgID)
		require.NoError(t, err)
		assert.Nil(t, visible)
	})
}

func TestAIModelRepo_GetVisibleByID_PropagatesContextError(t *testing.T) {
	db := setupAIModelRepoTestDB(t)
	repo := NewAIModelRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	model, err := repo.GetVisibleByID(ctx, 1, 1, 1)

	assert.Nil(t, model)
	require.Error(t, err)
}

func newAIModel(orgID, userID *int64) *aimodel.AIModel {
	return &aimodel.AIModel{
		OrganizationID: orgID,
		UserID:         userID,
		Name:           "Test model",
		ProviderType:   aimodel.ProviderTypeOpenAI,
		Model:          "gpt-test",
		IsEnabled:      true,
	}
}

func setupAIModelRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE ai_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER,
			user_id INTEGER,
			name TEXT NOT NULL,
			provider_type TEXT NOT NULL,
			model TEXT NOT NULL,
			base_url TEXT NOT NULL DEFAULT '',
			encrypted_credentials TEXT NOT NULL DEFAULT '',
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			token_budget INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}

func pointerTo(value int64) *int64 { return &value }
