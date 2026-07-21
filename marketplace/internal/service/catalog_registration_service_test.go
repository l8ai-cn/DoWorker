package service

import (
	"context"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/catalog"
	"github.com/stretchr/testify/require"
)

func TestCatalogRegistrationActivatesOnlyValidatedVersion(t *testing.T) {
	repository := &catalogRepositoryStub{}
	registration := NewCatalogRegistrationService(repository)
	ctx := context.Background()

	item, err := registration.RegisterItem(ctx, RegisterCatalogItemCommand{
		PublisherID:          2,
		Slug:                 "listing-optimizer",
		ResourceType:         "application",
		Name:                 "商品优化应用",
		Summary:              "开箱即用",
		PlatformResourceType: "expert",
		PlatformResourceID:   18,
		ActorUserID:          14,
	})
	require.NoError(t, err)

	version, err := registration.RegisterVersion(ctx, RegisterCatalogVersionCommand{
		CatalogItemID:  item.CatalogItemID,
		Version:        "1.0.0",
		SourceRevision: "git-sha",
		ContentDigest:  strings.Repeat("a", 64),
		Manifest:       []byte(`{"schema_version":"1"}`),
		Compatibility:  []byte(`{"agents":["codex-cli"]}`),
		ActorUserID:    14,
	})
	require.NoError(t, err)
	require.Equal(t, catalog.ValidationPending, version.ValidationStatus)

	activated, err := registration.MarkVersionPassed(ctx, version.CatalogItemVersionID)
	require.NoError(t, err)
	require.Equal(t, catalog.ItemStatusActive, activated.ItemStatus)
	require.Equal(t, version.CatalogItemVersionID, activated.LatestVersionID)
}

type catalogRepositoryStub struct {
	item    *catalog.Item
	version *catalog.Version
}

func (r *catalogRepositoryStub) CreateCatalogItem(
	_ context.Context,
	item *catalog.Item,
) (int64, error) {
	state := catalog.ItemState{
		ID:                      41,
		PublisherID:             item.PublisherID(),
		Slug:                    item.Slug().String(),
		ResourceType:            item.ResourceType(),
		Name:                    item.Name(),
		Summary:                 item.Summary(),
		PlatformResourceType:    item.PlatformResourceType(),
		PlatformResourceID:      item.PlatformResourceID(),
		CreatedByPlatformUserID: item.CreatedByPlatformUserID(),
		Status:                  catalog.ItemStatusDraft,
	}
	r.item, _ = catalog.RestoreItem(state)
	return 41, nil
}

func (r *catalogRepositoryStub) CreateCatalogVersion(
	_ context.Context,
	version *catalog.Version,
) (int64, error) {
	state := catalog.VersionState{
		ID:                      51,
		CatalogItemID:           version.CatalogItemID(),
		Version:                 version.Version(),
		SourceRevision:          version.SourceRevision(),
		ContentDigest:           version.ContentDigest(),
		Manifest:                version.Manifest(),
		Compatibility:           version.Compatibility(),
		ValidationStatus:        catalog.ValidationPending,
		CreatedByPlatformUserID: version.CreatedByPlatformUserID(),
	}
	r.version, _ = catalog.RestoreVersion(state)
	return 51, nil
}

func (r *catalogRepositoryStub) GetCatalogItem(context.Context, int64) (*catalog.Item, error) {
	return r.item, nil
}

func (r *catalogRepositoryStub) GetCatalogVersion(context.Context, int64) (*catalog.Version, error) {
	return r.version, nil
}

func (r *catalogRepositoryStub) ActivateCatalogVersion(
	_ context.Context,
	item *catalog.Item,
	version *catalog.Version,
) error {
	r.item = item
	r.version = version
	return nil
}
