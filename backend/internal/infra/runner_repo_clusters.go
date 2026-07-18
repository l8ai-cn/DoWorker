package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/executioncluster"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"gorm.io/gorm/clause"
)

func (r *runnerRepository) EnsureLocalClusterID(ctx context.Context, orgID int64) (int64, error) {
	defaults := []executioncluster.Cluster{
		{
			OrganizationID: orgID,
			Slug:           slugkit.Slug(executioncluster.KindOnline),
			Name:           "Online cluster",
			Kind:           executioncluster.KindOnline,
			Status:         executioncluster.StatusPending,
		},
		{
			OrganizationID: orgID,
			Slug:           slugkit.Slug(executioncluster.KindLocal),
			Name:           "Local cluster",
			Kind:           executioncluster.KindLocal,
			Status:         executioncluster.StatusPending,
		},
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "organization_id"}, {Name: "slug"}},
		DoNothing: true,
	}).Create(&defaults).Error; err != nil {
		return 0, err
	}

	var cluster executioncluster.Cluster
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, executioncluster.KindLocal).
		First(&cluster).Error; err != nil {
		return 0, err
	}
	return cluster.ID, nil
}
