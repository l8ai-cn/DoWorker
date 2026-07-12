package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/gorm"
)

type OrganizationApplicationsRepository struct {
	db *gorm.DB
}

func NewOrganizationApplicationsRepository(
	db *gorm.DB,
) *OrganizationApplicationsRepository {
	return &OrganizationApplicationsRepository{db: db}
}

func (r *OrganizationApplicationsRepository) ListOrganizationApplications(
	ctx context.Context,
	organizationID int64,
) ([]service.OrganizationApplication, error) {
	var rows []organizationApplicationRow
	result := r.db.WithContext(ctx).Raw(organizationApplicationsQuery, organizationID).Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}
	items := make([]service.OrganizationApplication, 0, len(rows))
	for _, row := range rows {
		var outcomes []string
		if err := json.Unmarshal(row.Outcomes, &outcomes); err != nil {
			return nil, err
		}
		items = append(items, service.OrganizationApplication{
			InstallationID: row.InstallationID,
			MarketSlug:     row.MarketSlug,
			ListingSlug:    row.ListingSlug,
			DisplayName:    row.DisplayName,
			Tagline:        row.Tagline,
			ResourceType:   row.ResourceType,
			Outcomes:       outcomes,
			RuntimeRef:     row.RuntimeRef,
			Status:         row.Status,
			InstalledAt:    row.InstalledAt,
		})
	}
	return items, nil
}

type organizationApplicationRow struct {
	InstallationID string
	MarketSlug     string
	ListingSlug    string
	DisplayName    string
	Tagline        string
	ResourceType   string
	Outcomes       []byte
	RuntimeRef     string
	Status         string
	InstalledAt    time.Time
}

const organizationApplicationsQuery = `
SELECT i.id::text AS installation_id, m.slug AS market_slug, l.slug AS listing_slug,
  lv.display_name, lv.tagline, ci.resource_type, lv.outcomes,
  COALESCE(i.runtime_ref, '') AS runtime_ref, i.status, i.created_at AS installed_at
FROM marketplace.marketplace_installations i
JOIN marketplace.marketplaces m ON m.id = i.marketplace_id
JOIN marketplace.marketplace_listings l
  ON l.id = i.listing_id AND l.marketplace_id = i.marketplace_id
JOIN marketplace.marketplace_listing_versions lv
  ON lv.id = i.listing_version_id AND lv.listing_id = i.listing_id
JOIN marketplace.marketplace_catalog_items ci ON ci.id = l.catalog_item_id
WHERE i.target_platform_org_id = ?
  AND i.status IN ('installing', 'verifying', 'active')
ORDER BY i.created_at DESC, i.id DESC
`
