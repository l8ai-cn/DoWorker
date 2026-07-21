package service

import (
	"context"
	"testing"
	"time"

	marketdomain "github.com/l8ai-cn/agentcloud/marketplace/internal/domain/market"
	"github.com/stretchr/testify/require"
)

func TestCreateMarketBindsPrimaryDomainAndStartsConfiguring(t *testing.T) {
	repository := &marketConsoleRepositoryStub{}
	lifecycle, err := NewMarketLifecycleService(repository, "Markets.Example.com")
	require.NoError(t, err)

	created, err := lifecycle.CreateMarket(context.Background(), CreateMarketCommand{
		Slug:        "commerce-market",
		Name:        "跨境电商市场",
		Summary:     "开箱即用的 AI 应用",
		OwnerOrgID:  9,
		ActorUserID: 14,
	})
	require.NoError(t, err)
	require.Equal(t, "commerce-market.markets.example.com", repository.primaryHost)
	require.Equal(t, marketdomain.StatusConfiguring, repository.createdMarket.Status())
	require.Equal(t, "commerce-market", created.Slug)
	require.Equal(t, int64(1), created.Revision)
}

func TestCreateAndPublishSpaceUsesRevision(t *testing.T) {
	repository := &marketConsoleRepositoryStub{
		market: mustRestoreMarket(t, marketdomain.State{
			ID:                      42,
			Slug:                    "commerce-market",
			Name:                    "跨境电商市场",
			Summary:                 "开箱即用的 AI 应用",
			Status:                  marketdomain.StatusConfiguring,
			Visibility:              "private",
			OwnerPlatformOrgID:      9,
			CreatedByPlatformUserID: 14,
			Revision:                1,
		}),
	}
	lifecycle, err := NewMarketLifecycleService(repository, "markets.example.com")
	require.NoError(t, err)
	space, err := lifecycle.CreateSpace(context.Background(), CreateSpaceCommand{
		MarketSlug:  "commerce-market",
		Slug:        "operations",
		Name:        "运营",
		Summary:     "运营应用",
		ActorUserID: 14,
	})
	require.NoError(t, err)
	repository.space = repository.createdSpace

	published, err := lifecycle.PublishSpace(context.Background(), PublishSpaceCommand{
		MarketSlug:       "commerce-market",
		SpaceSlug:        "operations",
		ExpectedRevision: space.Revision,
		PublishedAt:      time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, marketdomain.SpaceStatusPublished, repository.savedSpace.Status())
	require.Equal(t, int64(2), published.Revision)
}

func TestPublishSpaceRejectsNonConfigurableMarket(t *testing.T) {
	repository := &marketConsoleRepositoryStub{
		market: mustRestoreMarket(t, marketdomain.State{
			ID:                      42,
			Slug:                    "commerce-market",
			Name:                    "跨境电商市场",
			Summary:                 "开箱即用的 AI 应用",
			Status:                  marketdomain.StatusSuspended,
			Visibility:              "private",
			OwnerPlatformOrgID:      9,
			CreatedByPlatformUserID: 14,
			Revision:                3,
		}),
	}
	lifecycle, err := NewMarketLifecycleService(repository, "markets.example.com")
	require.NoError(t, err)

	_, err = lifecycle.PublishSpace(context.Background(), PublishSpaceCommand{
		MarketSlug:       "commerce-market",
		SpaceSlug:        "operations",
		ExpectedRevision: 1,
		PublishedAt:      time.Now().UTC(),
	})
	require.ErrorIs(t, err, ErrMarketNotConfigurable)
}

func mustRestoreMarket(t *testing.T, state marketdomain.State) *marketdomain.Market {
	t.Helper()
	item, err := marketdomain.Restore(state)
	require.NoError(t, err)
	return item
}

type marketConsoleRepositoryStub struct {
	market        *marketdomain.Market
	space         *marketdomain.Space
	createdMarket *marketdomain.Market
	createdSpace  *marketdomain.Space
	savedSpace    *marketdomain.Space
	primaryHost   string
}

func (r *marketConsoleRepositoryStub) CreateMarketWithDomain(
	_ context.Context,
	item *marketdomain.Market,
	primaryHost string,
) error {
	item.ID = 42
	r.createdMarket = item
	r.primaryHost = primaryHost
	return nil
}

func (r *marketConsoleRepositoryStub) GetMarketBySlug(
	context.Context,
	string,
) (*marketdomain.Market, error) {
	return r.market, nil
}

func (r *marketConsoleRepositoryStub) CreateSpace(
	_ context.Context,
	item *marketdomain.Space,
) error {
	item.ID = 11
	r.createdSpace = item
	return nil
}

func (r *marketConsoleRepositoryStub) GetSpace(
	context.Context,
	int64,
	string,
) (*marketdomain.Space, error) {
	return r.space, nil
}

func (r *marketConsoleRepositoryStub) SaveSpace(
	_ context.Context,
	item *marketdomain.Space,
	_ int64,
) error {
	r.savedSpace = item
	return nil
}
