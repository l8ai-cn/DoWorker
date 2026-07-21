package infra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
)

func TestExtensionRepo_BatchUpsertMcpMarketItems_MatchesRegistryPartialUniqueIndex(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX idx_mcp_market_items_registry_name
		ON mcp_market_items(registry_name) WHERE registry_name IS NOT NULL
	`).Error)

	repo := NewExtensionRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.BatchUpsertMcpMarketItems(ctx, []*extension.McpMarketItem{{
		Slug:         "registry-item",
		Name:         "Original",
		RegistryName: "io.example/registry-item",
		Source:       extension.McpSourceRegistry,
	}}))
	require.NoError(t, repo.BatchUpsertMcpMarketItems(ctx, []*extension.McpMarketItem{{
		Slug:         "registry-item-updated",
		Name:         "Updated",
		RegistryName: "io.example/registry-item",
		Source:       extension.McpSourceRegistry,
	}}))

	var items []extension.McpMarketItem
	require.NoError(t, db.Where("registry_name = ?", "io.example/registry-item").Find(&items).Error)
	require.Len(t, items, 1)
	require.Equal(t, "Updated", items[0].Name)
}
