package expert

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarketInstallCleanupSurvivesCanceledRequestContext(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	published := fixture.publishCurrentSource(t)
	fixture.store.createErr = errors.New("create failed")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)

	require.EqualError(t, err, "create failed")
	require.Len(t, fixture.snapshots.deleteContexts, 1)
	require.NoError(t, fixture.snapshots.deleteErrors[0])
	require.Empty(t, fixture.snapshots.created)
}
