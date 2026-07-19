package expert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/stretchr/testify/require"
)

func TestMarketInstallRevalidatesPublishedReleaseInsideApplicationLock(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	published := fixture.publishCurrentSource(t)
	fixture.locker.applicationHook = func() {
		release := fixture.market.releases[published.Release.ID]
		release.Status = expertmarket.ReleaseStatusWithdrawn
		fixture.market.releases[release.ID] = release
		application := fixture.market.applications[published.Application.ID]
		application.LatestPublishedReleaseID = nil
		fixture.market.applications[application.ID] = application
	}

	_, _, err := fixture.service.InstallPublishedMarketApplication(
		context.Background(),
		InstallMarketApplicationRequest{
			OrganizationID: 42, OrganizationSlug: "target-org",
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)

	require.ErrorIs(t, err, ErrMarketApplicationNotFound)
	require.Empty(t, fixture.snapshots.created)
}

func TestMarketUpgradeSerializesConsumerEdit(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	v1 := fixture.publishCurrentSource(t)
	installed, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID: 42, OrganizationSlug: "target-org", UserID: 501, ModelResourceID: 301,
			MarketSlug: string(v1.Application.Slug),
		},
	)
	require.NoError(t, err)
	fixture.source.Name = "Video Production Expert V2"
	require.NoError(t, fixture.store.Update(ctx, fixture.source))
	submission, err := fixture.service.SubmitMarketApplication(
		ctx, fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{ReviewerUserID: 99, ReleaseID: submission.Release.ID},
	)
	require.NoError(t, err)
	fixture.store.marketUpdateStarted = make(chan struct{})
	fixture.store.marketUpdateContinue = make(chan struct{})
	upgradeDone := make(chan error, 1)
	go func() {
		_, _, upgradeErr := fixture.service.UpgradeMarketApplication(
			ctx,
			UpgradeMarketApplicationRequest{OrganizationID: 42, OrganizationSlug: "target-org", UserID: 501, ExpertID: installed.ID},
		)
		upgradeDone <- upgradeErr
	}()
	<-fixture.store.marketUpdateStarted
	localName := "My Local Video Expert"
	editDone := make(chan error, 1)
	go func() {
		_, editErr := fixture.service.Update(ctx, &UpdateExpertRequest{
			OrganizationID: 42, ExpertID: installed.ID,
			Name: &localName,
		})
		editDone <- editErr
	}()
	select {
	case err := <-editDone:
		t.Fatalf("consumer edit bypassed market lock: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(fixture.store.marketUpdateContinue)
	require.NoError(t, <-upgradeDone)
	require.NoError(t, <-editDone)
	stored, err := fixture.store.GetByID(ctx, 42, installed.ID)
	require.NoError(t, err)
	require.Equal(t, localName, stored.Name)
	require.Equal(t, submission.Release.ID, *stored.SourceMarketReleaseID)
}

func TestMarketUpgradeCleanupSurvivesRequestCancellation(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	v1 := fixture.publishCurrentSource(t)
	installed, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID: 42, OrganizationSlug: "target-org", UserID: 501, ModelResourceID: 301,
			MarketSlug: string(v1.Application.Slug),
		},
	)
	require.NoError(t, err)
	fixture.source.Name = "Video Production Expert V2"
	require.NoError(t, fixture.store.Update(ctx, fixture.source))
	submission, err := fixture.service.SubmitMarketApplication(
		ctx, fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{ReviewerUserID: 99, ReleaseID: submission.Release.ID},
	)
	require.NoError(t, err)
	fixture.store.beforeMarketUpdate = cancel
	fixture.store.updateErr = context.Canceled

	_, _, err = fixture.service.UpgradeMarketApplication(
		ctx,
		UpgradeMarketApplicationRequest{OrganizationID: 42, OrganizationSlug: "target-org", UserID: 501, ExpertID: installed.ID},
	)

	require.True(t, errors.Is(err, context.Canceled))
	require.Len(t, fixture.snapshots.created, 1)
	require.NotEmpty(t, fixture.snapshots.deleteContexts)
	require.NoError(t, fixture.snapshots.deleteErrors[0])
}
