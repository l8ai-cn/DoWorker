package infra

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAIResourceRepositoryPostgresRoundTrip(t *testing.T) {
	db := openAIResourcePostgresTestDB(t)
	repo := NewAIResourceRepository(db)
	ctx := context.Background()

	connection := &airesource.Connection{
		OwnerScope: airesource.OwnerScopeUser, OwnerID: 1,
		Identifier: slugkit.Slug("postgres-openai"), ProviderKey: slugkit.Slug("openai"),
		Name: "Postgres OpenAI", CredentialsEncrypted: "encrypted",
		ConfiguredFields: []string{"api-key"}, Status: airesource.ConnectionStatusValid,
		IsEnabled: true, CreatedBy: 1,
	}
	require.NoError(t, repo.CreateConnection(ctx, connection))
	t.Cleanup(func() {
		assert.NoError(t, repo.DeleteConnection(context.Background(), connection.ID, connection.Revision))
	})
	resource := &airesource.ModelResource{
		ProviderConnectionID: connection.ID, Identifier: slugkit.Slug("postgres-gpt"),
		ModelID: "gpt-5.5", DisplayName: "GPT",
		Modalities:   []airesource.Modality{airesource.ModalityChat, airesource.ModalityImage},
		Capabilities: []airesource.Capability{airesource.CapabilityTextGeneration},
		Status:       airesource.ConnectionStatusValid, IsEnabled: true,
	}
	require.NoError(t, repo.CreateResource(ctx, resource))
	require.NoError(t, repo.SetDefault(ctx, resource.ID, airesource.ModalityChat))

	loaded, err := repo.GetConnectionByID(ctx, connection.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"api-key"}, loaded.ConfiguredFields)
	effective, err := repo.ListEffective(ctx, 1, 0, nil)
	require.NoError(t, err)
	require.Len(t, effective, 1)
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat, airesource.ModalityImage}, effective[0].Modalities)
	assert.Equal(t, []airesource.Modality{airesource.ModalityChat}, effective[0].DefaultModalities)

	orphan := *connection
	orphan.ID = 0
	orphan.Identifier = slugkit.Slug("postgres-orphan")
	orphan.OwnerID = 999
	require.Error(t, repo.CreateConnection(ctx, &orphan))
	require.Error(t, db.Exec("DELETE FROM users WHERE id = ?", 1).Error)
	resource.Modalities = []airesource.Modality{airesource.ModalityImage}
	require.Error(t, repo.SaveResource(ctx, resource))
	connection.OwnerScope = airesource.OwnerScopeOrg
	connection.OwnerID = 10
	require.Error(t, repo.SaveConnection(ctx, connection))
}

func TestAIResourceRepositoryPostgresOwnerLock(t *testing.T) {
	db := openAIResourcePostgresTestDB(t)
	t.Cleanup(func() {
		assert.NoError(t, db.Exec("DELETE FROM provider_connections WHERE identifier = ?", "postgres-concurrent").Error)
	})
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	insertTx, err := sqlDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = insertTx.Rollback() }()
	_, err = insertTx.ExecContext(ctx, `INSERT INTO provider_connections
		(owner_scope, owner_id, identifier, provider_key, name, created_by)
		VALUES ('org', 10, 'postgres-concurrent', 'openai', 'Concurrent', 1)`)
	require.NoError(t, err)

	deleteConnection, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	defer deleteConnection.Close()
	var deletePID int
	require.NoError(t, deleteConnection.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&deletePID))
	deleteDone := make(chan error, 1)
	go func() {
		_, deleteErr := deleteConnection.ExecContext(ctx, "DELETE FROM organizations WHERE id = 10")
		deleteDone <- deleteErr
	}()

	require.NoError(t, waitForPostgresLock(ctx, db, deletePID, deleteDone))
	require.NoError(t, insertTx.Commit())
	require.Error(t, <-deleteDone)
	assertPostgresCount(t, db, "organizations", "id = 10", 1)
	assertPostgresCount(t, db, "provider_connections", "owner_scope = 'org' AND owner_id = 10", 1)
}

func openAIResourcePostgresTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AI_RESOURCE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("AI_RESOURCE_POSTGRES_TEST_DSN is not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return db
}

func waitForPostgresLock(ctx context.Context, db *gorm.DB, pid int, done <-chan error) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			return fmt.Errorf("owner delete completed before insert commit: %v", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var waiting bool
			err := db.Raw(`SELECT EXISTS (
				SELECT 1 FROM pg_stat_activity
				WHERE pid = ? AND wait_event_type = 'Lock'
			)`, pid).Scan(&waiting).Error
			if err != nil {
				return err
			}
			if waiting {
				return nil
			}
		}
	}
}

func assertPostgresCount(t *testing.T, db *gorm.DB, table, condition string, expected int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Table(table).Where(condition).Count(&count).Error)
	assert.Equal(t, expected, count)
}
