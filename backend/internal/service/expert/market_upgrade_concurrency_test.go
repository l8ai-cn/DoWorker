package expert

import (
	"context"
	"sync"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/require"
)

func TestMarketUpgradeSerializesConcurrentRequests(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fixture.service.gitops = gitops.NewFake("am-experts")
	ctx := context.Background()
	v1 := fixture.publishCurrentSource(t)
	installed, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(v1.Application.Slug),
		},
	)
	require.NoError(t, err)

	fixture.source.Name = "Video Production Expert V2"
	require.NoError(t, fixture.store.Update(ctx, fixture.source))
	submission, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{
			ReviewerUserID: 99,
			ReleaseID:      submission.Release.ID,
		},
	)
	require.NoError(t, err)

	start := make(chan struct{})
	results := make(chan bool, 2)
	errors := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, changed, upgradeErr := fixture.service.UpgradeMarketApplication(
				ctx,
				UpgradeMarketApplicationRequest{
					OrganizationID: 42,
					UserID:         501,
					ExpertID:       installed.ID,
				},
			)
			results <- changed
			errors <- upgradeErr
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	close(errors)

	for upgradeErr := range errors {
		require.NoError(t, upgradeErr)
	}
	var changedCount int
	for changed := range results {
		if changed {
			changedCount++
		}
	}
	require.Equal(t, 1, changedCount)
	require.Len(t, fixture.snapshots.created, 2)
	require.Equal(t, 3, fixture.locker.calls)
}
