package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ExpertCatalogSynchronizer struct {
	db *gorm.DB
}

type publishedExpertRelease struct {
	ApplicationID           int64
	ReleaseID               int64
	PublisherOrganizationID int64
	PublisherUserID         int64
	ReviewerUserID          int64
	Version                 int
	Slug                    string
	Name                    string
	Summary                 string
	Description             string
	PublisherSlug           string
	PublisherName           string
	IsOperatorOwned         bool
	Featured                bool
	Tags                    pq.StringArray `gorm:"type:text[]"`
	Outcomes                pq.StringArray `gorm:"type:text[]"`
	ExpertSnapshot          []byte
	WorkerSpecSnapshot      []byte
	PublishedAt             time.Time
}

func NewExpertCatalogSynchronizer(db *gorm.DB) *ExpertCatalogSynchronizer {
	return &ExpertCatalogSynchronizer{db: db}
}

func (s *ExpertCatalogSynchronizer) Sync(ctx context.Context) (int, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("expert catalog synchronizer database is required")
	}
	releases, err := s.loadLatestPublishedReleases(ctx)
	if err != nil {
		return 0, err
	}
	for _, release := range releases {
		if release.ReleaseID == 0 {
			if err := s.removeListing(ctx, release.ApplicationID); err != nil {
				return 0, fmt.Errorf("remove expert listing %s: %w", release.Slug, err)
			}
			continue
		}
		payload, err := buildExpertCatalogPayload(release)
		if err != nil {
			return 0, fmt.Errorf("build expert listing %s: %w", release.Slug, err)
		}
		if err := s.publishListing(ctx, release, payload); err != nil {
			return 0, fmt.Errorf("publish expert listing %s: %w", release.Slug, err)
		}
	}
	if err := s.removeOrphanedListings(ctx); err != nil {
		return 0, fmt.Errorf("remove orphaned expert listings: %w", err)
	}
	return len(releases), nil
}

func (s *ExpertCatalogSynchronizer) loadLatestPublishedReleases(
	ctx context.Context,
) ([]publishedExpertRelease, error) {
	var rows []publishedExpertRelease
	err := s.db.WithContext(ctx).Raw(`
SELECT a.id AS application_id, COALESCE(r.id, 0) AS release_id,
  a.publisher_organization_id, COALESCE(r.publisher_user_id, 0) AS publisher_user_id,
  COALESCE(r.reviewer_user_id, 0) AS reviewer_user_id, COALESCE(r.version, 0) AS version,
  a.slug, COALESCE(r.name, '') AS name, COALESCE(r.summary, '') AS summary,
  COALESCE(r.description, '') AS description, o.slug AS publisher_slug,
  o.name AS publisher_name, a.is_operator_owned, COALESCE(r.featured, FALSE) AS featured,
  COALESCE(r.tags, '{}') AS tags, COALESCE(r.outcomes, '{}') AS outcomes,
  COALESCE(r.expert_snapshot, '{}'::jsonb) AS expert_snapshot,
  COALESCE(r.worker_spec_snapshot, '{}'::jsonb) AS worker_spec_snapshot,
  COALESCE(r.published_at, 'epoch'::timestamptz) AS published_at
FROM expert_market_applications a
JOIN organizations o ON o.id = a.publisher_organization_id
LEFT JOIN expert_market_releases r ON r.id = a.latest_published_release_id
ORDER BY a.id
`).Scan(&rows).Error
	return rows, err
}

func (s *ExpertCatalogSynchronizer) removeOrphanedListings(
	ctx context.Context,
) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE marketplace.marketplace_listings l
SET status = 'removed', visibility = 'hidden',
  revision = l.revision + 1, updated_at = NOW()
FROM marketplace.marketplace_catalog_items ci
WHERE l.catalog_item_id = ci.id
  AND ci.platform_resource_type = 'expert'
  AND ci.platform_resource_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM expert_market_applications a
    WHERE a.id = ci.platform_resource_id
  )
  AND l.status <> 'removed'
`).Error
}
