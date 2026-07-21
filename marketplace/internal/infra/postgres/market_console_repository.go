package postgres

import (
	"context"
	"time"

	marketdomain "github.com/l8ai-cn/agentcloud/marketplace/internal/domain/market"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

type MarketConsoleRepository struct {
	db *gorm.DB
}

func NewMarketConsoleRepository(db *gorm.DB) *MarketConsoleRepository {
	return &MarketConsoleRepository{db: db}
}

func (r *MarketConsoleRepository) CreateMarketWithDomain(
	ctx context.Context,
	item *marketdomain.Market,
	primaryHost string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var result struct{ ID int64 }
		if err := tx.Raw(`
INSERT INTO marketplace.marketplaces
  (slug, name, summary, status, visibility, owner_platform_org_id,
   created_by_platform_user_id, revision)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id
`, item.Slug().String(), item.Name, item.Summary, item.Status(), item.Visibility,
			item.OwnerPlatformOrgID, item.CreatedByPlatformUserID, item.Revision()).
			Scan(&result).Error; err != nil {
			return err
		}
		item.ID = result.ID
		return tx.Exec(`
INSERT INTO marketplace.marketplace_domains
  (marketplace_id, host, kind, status, verification_token, is_primary, verified_at)
VALUES (?, ?, 'platform', 'active', ?, TRUE, NOW())
`, item.ID, primaryHost, "platform:"+primaryHost).Error
	})
}

func (r *MarketConsoleRepository) GetMarketBySlug(
	ctx context.Context,
	slug string,
) (*marketdomain.Market, error) {
	var row struct {
		ID                      int64
		Slug                    string
		Name                    string
		Summary                 string
		Status                  marketdomain.Status
		Visibility              string
		OwnerPlatformOrgID      int64
		CreatedByPlatformUserID int64
		Revision                int64
	}
	result := r.db.WithContext(ctx).Raw(`
SELECT id, slug, name, summary, status, visibility, owner_platform_org_id,
  created_by_platform_user_id, revision
FROM marketplace.marketplaces WHERE slug = ? LIMIT 1
`, slug).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrMarketNotFound
	}
	return marketdomain.Restore(marketdomain.State{
		ID:                      row.ID,
		Slug:                    row.Slug,
		Name:                    row.Name,
		Summary:                 row.Summary,
		Status:                  row.Status,
		Visibility:              row.Visibility,
		OwnerPlatformOrgID:      row.OwnerPlatformOrgID,
		CreatedByPlatformUserID: row.CreatedByPlatformUserID,
		Revision:                row.Revision,
	})
}

func (r *MarketConsoleRepository) CreateSpace(
	ctx context.Context,
	item *marketdomain.Space,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockConfigurableMarket(tx, item.MarketplaceID); err != nil {
			return err
		}
		var result struct{ ID int64 }
		if err := tx.Raw(`
INSERT INTO marketplace.marketplace_spaces
  (marketplace_id, slug, name, summary, status, created_by_platform_user_id, revision)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id
`, item.MarketplaceID, item.Slug().String(), item.Name, item.Summary, item.Status(),
			item.CreatedByPlatformUserID, item.Revision()).Scan(&result).Error; err != nil {
			return err
		}
		item.ID = result.ID
		return nil
	})
}

func (r *MarketConsoleRepository) GetSpace(
	ctx context.Context,
	marketplaceID int64,
	slug string,
) (*marketdomain.Space, error) {
	var row marketdomain.SpaceState
	result := r.db.WithContext(ctx).Raw(`
SELECT id, marketplace_id, slug, name, summary, status, revision,
  created_by_platform_user_id, published_at
FROM marketplace.marketplace_spaces
WHERE marketplace_id = ? AND slug = ? LIMIT 1
`, marketplaceID, slug).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrSpaceNotFound
	}
	return marketdomain.RestoreSpace(row)
}

func (r *MarketConsoleRepository) SaveSpace(
	ctx context.Context,
	item *marketdomain.Space,
	expectedRevision int64,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockConfigurableMarket(tx, item.MarketplaceID); err != nil {
			return err
		}
		result := tx.Exec(`
UPDATE marketplace.marketplace_spaces
SET status = ?, published_at = ?, revision = revision + 1, updated_at = ?
WHERE id = ? AND revision = ?
`, item.Status(), item.PublishedAt(), time.Now().UTC(), item.ID, expectedRevision)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return service.ErrRevisionConflict
		}
		return nil
	})
}
