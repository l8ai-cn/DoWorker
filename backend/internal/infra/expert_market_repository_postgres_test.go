package infra

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExpertMarketRepositoryMapsPostgresForeignKeyConflicts(t *testing.T) {
	db := openExpertMarketPostgresTestDB(t)
	repo := NewExpertMarketRepository(db)
	ctx := context.Background()

	app := expertmarket.Application{
		Slug:                    slugkit.Slug("postgres-video"),
		PublisherOrganizationID: 1,
		PublisherUserID:         1,
	}
	require.NoError(t, repo.CreateApplication(ctx, &app))

	for name, release := range map[string]expertmarket.Release{
		"foreign expert":      postgresTestRelease(app.ID, 9002, 1),
		"missing application": postgresTestRelease(9999, 9001, 1),
	} {
		t.Run(name, func(t *testing.T) {
			err := repo.CreateRelease(ctx, &release)
			require.ErrorIs(t, err, expertmarket.ErrConflict)
			require.NotContains(t, err.Error(), "23503")
		})
	}
}

func TestExpertMarketPostgresConcurrentPublicationKeepsNewestLatest(t *testing.T) {
	db := openExpertMarketPostgresTestDB(t)
	repo := NewExpertMarketRepository(db)
	ctx := context.Background()
	app := expertmarket.Application{
		Slug:                    slugkit.Slug("concurrent-video"),
		PublisherOrganizationID: 1,
		PublisherUserID:         1,
	}
	require.NoError(t, repo.CreateApplication(ctx, &app))

	v1 := postgresTestRelease(app.ID, 9001, 1)
	v1.Status = expertmarket.ReleaseStatusPendingReview
	v2 := v1
	v2.Version = 2
	require.NoError(t, repo.CreateRelease(ctx, &v1))
	require.NoError(t, repo.CreateRelease(ctx, &v2))

	start := make(chan struct{})
	errs := make(chan error, 2)
	var workers sync.WaitGroup
	for _, releaseID := range []int64{v1.ID, v2.ID} {
		workers.Add(1)
		go func(id int64) {
			defer workers.Done()
			<-start
			errs <- repo.UpdateReleaseLifecycleAndLatest(
				ctx,
				app.ID,
				id,
				expertmarket.LifecycleUpdate{
					Status: expertmarket.ReleaseStatusPublished,
				},
			)
		}(releaseID)
	}
	close(start)
	workers.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	stored, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.LatestPublishedReleaseID)
	require.Equal(t, v2.ID, *stored.LatestPublishedReleaseID)
}

func TestExpertMarketPostgresWithdrawalRestoresPreviousRelease(t *testing.T) {
	db := openExpertMarketPostgresTestDB(t)
	repo := NewExpertMarketRepository(db)
	ctx := context.Background()
	app := expertmarket.Application{
		Slug:                    slugkit.Slug("withdraw-video"),
		PublisherOrganizationID: 1,
		PublisherUserID:         1,
	}
	require.NoError(t, repo.CreateApplication(ctx, &app))

	v1 := postgresTestRelease(app.ID, 9001, 1)
	v1.Status = expertmarket.ReleaseStatusPendingReview
	v2 := v1
	v2.Version = 2
	require.NoError(t, repo.CreateRelease(ctx, &v1))
	require.NoError(t, repo.CreateRelease(ctx, &v2))
	for _, releaseID := range []int64{v1.ID, v2.ID} {
		require.NoError(t, repo.UpdateReleaseLifecycleAndLatest(
			ctx,
			app.ID,
			releaseID,
			expertmarket.LifecycleUpdate{
				Status: expertmarket.ReleaseStatusPublished,
			},
		))
	}
	require.NoError(t, repo.WithdrawReleaseAndRefreshLatest(
		ctx,
		app.ID,
		v2.ID,
		expertmarket.LifecycleUpdate{
			Status: expertmarket.ReleaseStatusWithdrawn,
		},
	))

	stored, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Equal(t, v1.ID, *stored.LatestPublishedReleaseID)
}

func openExpertMarketPostgresTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}
	schema := "expert_market_repo_" + strings.ReplaceAll(
		time.Now().UTC().Format("20060102150405.000000000"),
		".",
		"",
	)
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	_, err = admin.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`)
	})

	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec(expertMarketRepositoryBaseDDL).Error)
	up, err := os.ReadFile("../../migrations/000208_expert_marketplace.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
	return db
}

func postgresTestRelease(
	applicationID, sourceExpertID, publisherOrganizationID int64,
) expertmarket.Release {
	release := testRelease(applicationID, 1, expertmarket.ReleaseStatusDraft)
	release.SourceExpertID = sourceExpertID
	release.PublisherOrganizationID = publisherOrganizationID
	return release
}

const expertMarketRepositoryBaseDDL = `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
CREATE TABLE experts (id BIGINT PRIMARY KEY, organization_id BIGINT NOT NULL);
INSERT INTO users(id) VALUES (1), (2);
INSERT INTO organizations(id) VALUES (1), (2);
INSERT INTO experts(id, organization_id) VALUES (9001, 1), (9002, 2);
`
