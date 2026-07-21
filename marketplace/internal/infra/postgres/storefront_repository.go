package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

type StorefrontRepository struct {
	db *gorm.DB
}

func NewStorefrontRepository(db *gorm.DB) *StorefrontRepository {
	return &StorefrontRepository{db: db}
}

func (r *StorefrontRepository) ResolveMarket(
	ctx context.Context,
	marketSlug string,
	host string,
) (service.MarketView, error) {
	var row service.MarketView
	result := r.db.WithContext(ctx).Raw(`
SELECT m.id AS marketplace_id, m.slug, m.name, m.summary, m.status, m.default_locale
FROM marketplace.marketplaces m
JOIN marketplace.marketplace_domains d
  ON d.marketplace_id = m.id
 AND d.host = ?
 AND d.status = 'active'
WHERE m.slug = ?
  AND m.status IN ('published', 'suspended')
  AND m.visibility = 'public'
LIMIT 1
`, host, marketSlug).Scan(&row)
	if result.Error != nil {
		return service.MarketView{}, result.Error
	}
	if row.MarketplaceID == 0 {
		return service.MarketView{}, service.ErrMarketNotFound
	}
	return row, nil
}
