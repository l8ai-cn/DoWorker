package infra

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/stretchr/testify/require"
)

func TestExpertMarketGenericLifecycleCannotPublish(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))

	publishedAt := time.Now().UTC()
	err := repo.UpdateReleaseLifecycle(ctx, release.ID, expertmarket.LifecycleUpdate{
		Status:      expertmarket.ReleaseStatusPublished,
		PublishedAt: &publishedAt,
	})
	require.ErrorIs(t, err, expertmarket.ErrPublicationRequiresLatestUpdate)

	stored, err := repo.GetReleaseByID(ctx, release.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPendingReview, stored.Status)
	require.Nil(t, stored.PublishedAt)
}

func TestExpertMarketGenericLifecycleCannotDemoteLatest(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))

	publishedAt := time.Now().UTC()
	require.NoError(t, repo.UpdateReleaseLifecycleAndLatest(
		ctx,
		app.ID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status:      expertmarket.ReleaseStatusPublished,
			PublishedAt: &publishedAt,
		},
	))

	withdrawnAt := publishedAt.Add(time.Hour)
	err := repo.UpdateReleaseLifecycle(ctx, release.ID, expertmarket.LifecycleUpdate{
		Status:      expertmarket.ReleaseStatusWithdrawn,
		WithdrawnAt: &withdrawnAt,
	})
	require.ErrorIs(t, err, expertmarket.ErrLatestReleaseStatusConflict)

	storedRelease, err := repo.GetReleaseByID(ctx, release.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPublished, storedRelease.Status)
	require.Nil(t, storedRelease.WithdrawnAt)
	storedApplication, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Equal(t, release.ID, *storedApplication.LatestPublishedReleaseID)
}

func TestExpertMarketDelayedOlderPublicationDoesNotRollbackLatest(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	v1 := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	v2 := testRelease(app.ID, 2, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &v1))
	require.NoError(t, repo.CreateRelease(ctx, &v2))

	publishedAt := time.Now().UTC()
	for _, release := range []*expertmarket.Release{&v2, &v1} {
		require.NoError(t, repo.UpdateReleaseLifecycleAndLatest(
			ctx,
			app.ID,
			release.ID,
			expertmarket.LifecycleUpdate{
				Status:      expertmarket.ReleaseStatusPublished,
				PublishedAt: &publishedAt,
			},
		))
	}

	storedV1, err := repo.GetReleaseByID(ctx, v1.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPublished, storedV1.Status)
	storedApplication, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Equal(t, v2.ID, *storedApplication.LatestPublishedReleaseID)
}

func TestExpertMarketWithdrawalRefreshesLatest(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	v1 := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	v2 := testRelease(app.ID, 2, expertmarket.ReleaseStatusPendingReview)
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
	withdrawnAt := time.Now().UTC()
	require.NoError(t, repo.WithdrawReleaseAndRefreshLatest(
		ctx,
		app.ID,
		v2.ID,
		expertmarket.LifecycleUpdate{
			Status:      expertmarket.ReleaseStatusWithdrawn,
			WithdrawnAt: &withdrawnAt,
		},
	))

	storedApplication, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Equal(t, v1.ID, *storedApplication.LatestPublishedReleaseID)
	storedV2, err := repo.GetReleaseByID(ctx, v2.ID)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusWithdrawn, storedV2.Status)
}

func TestExpertMarketWithdrawalClearsOnlyLatest(t *testing.T) {
	repo := newExpertMarketTestRepository(t)
	ctx := context.Background()
	app := testApplication("video-production", 10)
	require.NoError(t, repo.CreateApplication(ctx, &app))
	release := testRelease(app.ID, 1, expertmarket.ReleaseStatusPendingReview)
	require.NoError(t, repo.CreateRelease(ctx, &release))
	require.NoError(t, repo.UpdateReleaseLifecycleAndLatest(
		ctx,
		app.ID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status: expertmarket.ReleaseStatusPublished,
		},
	))
	require.NoError(t, repo.WithdrawReleaseAndRefreshLatest(
		ctx,
		app.ID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status: expertmarket.ReleaseStatusWithdrawn,
		},
	))

	storedApplication, err := repo.GetApplicationByID(ctx, app.ID)
	require.NoError(t, err)
	require.Nil(t, storedApplication.LatestPublishedReleaseID)
	require.ErrorIs(t, repo.WithdrawReleaseAndRefreshLatest(
		ctx,
		app.ID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status: expertmarket.ReleaseStatusRejected,
		},
	), expertmarket.ErrInvalidWithdrawalStatus)
}
