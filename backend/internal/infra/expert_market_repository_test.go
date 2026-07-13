package infra

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExpertMarketRepositoryApplications(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()

	first := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &first))
	second := testApplication("video-editing", 20)
	require.NoError(t, repo.CreateApplication(ctx, &second))

	byID, err := repo.GetApplicationByID(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first.Slug, byID.Slug)

	bySlug, err := repo.GetApplicationBySlug(ctx, string(second.Slug))
	require.NoError(t, err)
	require.Equal(t, second.ID, bySlug.ID)

	rows, total, err := repo.ListApplications(ctx, expertmarket.ApplicationListFilter{
		PublisherOrganizationID: ptrInt64(10),
		Limit:                   10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, first.ID, rows[0].ID)

	_, err = repo.GetApplicationBySlug(ctx, "missing")
	require.ErrorIs(t, err, expertmarket.ErrNotFound)

	duplicate := testApplication(string(first.Slug), 30)
	require.ErrorIs(t, repo.CreateApplication(ctx, &duplicate), expertmarket.ErrConflict)
}

func TestExpertMarketRepositoryReleases(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))

	first := testRelease(app.ID, 1, expertmarket.ReleaseStatusDraft)
	require.NoError(t, repo.CreateRelease(ctx, &first))
	second := testRelease(app.ID, 2, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &second))

	got, err := repo.GetReleaseByID(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first.ExpertSnapshot, got.ExpertSnapshot)

	status := expertmarket.ReleaseStatusPendingReview
	rows, total, err := repo.ListReleases(ctx, expertmarket.ReleaseListFilter{
		ApplicationID: &app.ID,
		Status:        &status,
		Limit:         10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, second.ID, rows[0].ID)

	duplicate := testRelease(app.ID, 1, expertmarket.ReleaseStatusDraft)
	require.ErrorIs(t, repo.CreateRelease(ctx, &duplicate), expertmarket.ErrConflict)

	_, err = repo.GetReleaseByID(ctx, 9999)
	require.ErrorIs(t, err, expertmarket.ErrNotFound)
}

func TestExpertMarketRepositoryUpdatesLifecycle(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))

	now := time.Now().UTC().Truncate(time.Second)
	reviewerID := int64(99)
	update := expertmarket.LifecycleUpdate{
		Status:         expertmarket.ReleaseStatusPublished,
		ReviewerUserID: &reviewerID,
		ReviewedAt:     &now,
		PublishedAt:    &now,
	}
	require.NoError(t, repo.UpdateReleaseLifecycle(ctx, release.ID, update))

	got, err := repo.GetReleaseByID(ctx, release.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPublished, got.Status)
	require.Equal(t, reviewerID, *got.ReviewerUserID)
	require.Equal(t, now, got.PublishedAt.UTC())

	withdrawnAt := now.Add(time.Hour)
	require.NoError(t, repo.UpdateReleaseLifecycle(ctx, release.ID, expertmarket.LifecycleUpdate{
		Status:      expertmarket.ReleaseStatusWithdrawn,
		WithdrawnAt: &withdrawnAt,
	}))
	got, err = repo.GetReleaseByID(ctx, release.ID)
	require.NoError(t, err)
	require.Equal(t, reviewerID, *got.ReviewerUserID)
	require.Equal(t, now, got.PublishedAt.UTC())
	require.Equal(t, withdrawnAt, got.WithdrawnAt.UTC())

	require.ErrorIs(t,
		repo.UpdateReleaseLifecycle(ctx, 9999, update),
		expertmarket.ErrNotFound,
	)
}

func TestExpertMarketRepositoryUpdatesLifecycleAndLatestAtomically(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	other := testApplication("video-editing", 20)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	require.NoError(t, repo.CreateApplication(ctx, &other))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))

	now := time.Now().UTC()
	update := expertmarket.LifecycleUpdate{
		Status:      expertmarket.ReleaseStatusPublished,
		PublishedAt: &now,
	}
	require.Error(t, repo.UpdateReleaseLifecycleAndLatest(ctx, other.ID, release.ID, update))

	got, err := repo.GetReleaseByID(ctx, release.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPendingReview, got.Status)

	require.NoError(t, repo.UpdateReleaseLifecycleAndLatest(ctx, app.ID, release.ID, update))
	application, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Equal(t, release.ID, *application.LatestPublishedReleaseID)
}

func TestExpertMarketRepositoryRejectsNonPublishedLatestUpdates(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))

	reviewerID := int64(99)
	reviewedAt := time.Now().UTC()
	for _, status := range []expertmarket.ReleaseStatus{
		expertmarket.ReleaseStatusDraft,
		expertmarket.ReleaseStatusPendingReview,
		expertmarket.ReleaseStatusRejected,
		expertmarket.ReleaseStatusWithdrawn,
	} {
		t.Run(string(status), func(t *testing.T) {
			err := repo.UpdateReleaseLifecycleAndLatest(
				ctx,
				app.ID,
				release.ID,
				expertmarket.LifecycleUpdate{
					Status:         status,
					ReviewerUserID: &reviewerID,
					ReviewedAt:     &reviewedAt,
				},
			)
			require.ErrorIs(t, err, expertmarket.ErrInvalidLatestReleaseStatus)

			storedRelease, getErr := repo.GetReleaseByID(ctx, release.ID)
			require.NoError(t, getErr)
			require.Equal(t, expertmarket.ReleaseStatusPendingReview, storedRelease.Status)
			require.Nil(t, storedRelease.ReviewerUserID)
			require.Nil(t, storedRelease.ReviewedAt)

			storedApplication, getErr := repo.GetApplicationByID(ctx, app.ID)
			require.NoError(t, getErr)
			require.Nil(t, storedApplication.LatestPublishedReleaseID)
		})
	}
}

func newExpertMarketTestRepository(t *testing.T) expertmarket.Repository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.Exec(`PRAGMA foreign_keys = ON`).Error)
	require.NoError(t, db.Exec(`
CREATE TABLE expert_market_applications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT NOT NULL UNIQUE,
	publisher_organization_id INTEGER NOT NULL,
	publisher_user_id INTEGER NOT NULL,
	is_operator_owned BOOLEAN NOT NULL DEFAULT false,
	latest_published_release_id INTEGER,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`).Error)
	require.NoError(t, db.Exec(`
CREATE TABLE expert_market_releases (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	application_id INTEGER NOT NULL,
	source_expert_id INTEGER NOT NULL,
	publisher_organization_id INTEGER NOT NULL,
	publisher_user_id INTEGER NOT NULL,
	version INTEGER NOT NULL,
	status TEXT NOT NULL,
	name TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	category TEXT NOT NULL DEFAULT '',
	icon TEXT NOT NULL DEFAULT '',
	tags TEXT NOT NULL DEFAULT '{}',
	outcomes TEXT NOT NULL DEFAULT '{}',
	featured BOOLEAN NOT NULL DEFAULT false,
	expert_snapshot BLOB NOT NULL,
	worker_spec_snapshot BLOB NOT NULL,
	skill_dependencies BLOB NOT NULL,
	reviewer_user_id INTEGER,
	rejection_reason TEXT,
	submitted_at DATETIME,
	reviewed_at DATETIME,
	published_at DATETIME,
	rejected_at DATETIME,
	withdrawn_at DATETIME,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(application_id, version),
	UNIQUE(application_id, id),
	FOREIGN KEY(application_id) REFERENCES expert_market_applications(id)
)`).Error)
	return NewExpertMarketRepository(db)
}

func testApplication(slug string, orgID int64) expertmarket.Application {
	return expertmarket.Application{
		Slug:                    slugkit.Slug(slug),
		PublisherOrganizationID: orgID,
		PublisherUserID:         1,
	}
}

func testRelease(appID int64, version int, status expertmarket.ReleaseStatus) expertmarket.Release {
	return expertmarket.Release{
		ApplicationID:           appID,
		SourceExpertID:          30,
		PublisherOrganizationID: 10,
		PublisherUserID:         1,
		Version:                 version,
		Status:                  status,
		Name:                    "Video Expert",
		Tags:                    []string{"video"},
		Outcomes:                []string{"render"},
		ExpertSnapshot:          json.RawMessage(`{"version":1}`),
		WorkerSpecSnapshot:      json.RawMessage(`{"version":1}`),
		SkillDependencies:       json.RawMessage(`[]`),
	}
}

func ptrInt64(value int64) *int64 {
	return &value
}
