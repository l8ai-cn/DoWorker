package postgres

import (
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/gorm"
)

func lockConfigurableMarket(tx *gorm.DB, marketplaceID int64) error {
	var row struct {
		ID     int64
		Status string
	}
	result := tx.Raw(`
SELECT id, status
FROM marketplace.marketplaces
WHERE id = ?
FOR UPDATE
`, marketplaceID).Scan(&row)
	if result.Error != nil {
		return result.Error
	}
	if row.ID == 0 {
		return service.ErrMarketNotFound
	}
	if row.Status != "configuring" && row.Status != "review" {
		return service.ErrMarketNotConfigurable
	}
	return nil
}
