package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/executioncluster"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type executionClusterRepository struct{ db *gorm.DB }

func NewExecutionClusterRepository(db *gorm.DB) executioncluster.Repository {
	return &executionClusterRepository{db: db}
}

func (r *executionClusterRepository) ListByOrganization(ctx context.Context, organizationID int64) ([]*executioncluster.Cluster, error) {
	var clusters []*executioncluster.Cluster
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("kind ASC").
		Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

func (r *executionClusterRepository) GetByIDAndOrganization(ctx context.Context, id, organizationID int64) (*executioncluster.Cluster, error) {
	var cluster executioncluster.Cluster
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		First(&cluster).Error; err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &cluster, nil
}

func (r *executionClusterRepository) EnsureDefaults(ctx context.Context, organizationID int64) ([]*executioncluster.Cluster, error) {
	defaults := []*executioncluster.Cluster{
		{OrganizationID: organizationID, Slug: slugkit.Slug("online"), Name: "线上集群", Kind: executioncluster.KindOnline, Status: executioncluster.StatusPending},
		{OrganizationID: organizationID, Slug: slugkit.Slug("local"), Name: "本地集群", Kind: executioncluster.KindLocal, Status: executioncluster.StatusPending},
	}
	for _, cluster := range defaults {
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "organization_id"}, {Name: "slug"}},
			DoNothing: true,
		}).Create(cluster).Error; err != nil {
			return nil, err
		}
	}
	return r.ListByOrganization(ctx, organizationID)
}
