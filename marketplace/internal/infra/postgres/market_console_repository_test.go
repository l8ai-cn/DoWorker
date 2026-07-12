package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMarketConsoleRepositoryCreatesAndPublishesSpace(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
	for _, migration := range []string{
		"000001_market_foundation.up.sql",
		"000002_core_catalog.up.sql",
		"000003_catalog_version_integrity.up.sql",
		"000004_console_revisions.up.sql",
	} {
		applyMigration(t, sqlDB, migration)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn))
	require.NoError(t, err)
	lifecycle, err := service.NewMarketLifecycleService(
		NewMarketConsoleRepository(db),
		"markets.example.com",
	)
	require.NoError(t, err)
	ctx := context.Background()

	market, err := lifecycle.CreateMarket(ctx, service.CreateMarketCommand{
		Slug:        "commerce-market",
		Name:        "跨境电商市场",
		Summary:     "开箱即用的 AI 应用",
		OwnerOrgID:  9,
		ActorUserID: 14,
	})
	require.NoError(t, err)
	require.Positive(t, market.MarketplaceID)

	space, err := lifecycle.CreateSpace(ctx, service.CreateSpaceCommand{
		MarketSlug:  "commerce-market",
		Slug:        "operations",
		Name:        "运营",
		Summary:     "运营应用",
		ActorUserID: 14,
	})
	require.NoError(t, err)

	publishedAt := time.Now().UTC()
	published, err := lifecycle.PublishSpace(ctx, service.PublishSpaceCommand{
		MarketSlug:       "commerce-market",
		SpaceSlug:        "operations",
		ExpectedRevision: space.Revision,
		PublishedAt:      publishedAt,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), published.Revision)

	_, err = lifecycle.PublishSpace(ctx, service.PublishSpaceCommand{
		MarketSlug:       "commerce-market",
		SpaceSlug:        "operations",
		ExpectedRevision: space.Revision,
		PublishedAt:      publishedAt,
	})
	require.ErrorIs(t, err, service.ErrRevisionConflict)

	var domainCount int64
	require.NoError(t, db.Raw(`
SELECT COUNT(*) FROM marketplace.marketplace_domains
WHERE marketplace_id = ? AND host = ? AND status = 'active' AND is_primary
	`, market.MarketplaceID, "commerce-market.markets.example.com").Scan(&domainCount).Error)
	require.Equal(t, int64(1), domainCount)
}
