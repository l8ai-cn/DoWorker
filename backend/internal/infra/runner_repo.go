package infra

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/grant"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"gorm.io/gorm"
)

var _ runner.RunnerRepository = (*runnerRepository)(nil)

type runnerRepository struct{ db *gorm.DB }

func NewRunnerRepository(db *gorm.DB) runner.RunnerRepository {
	return &runnerRepository{db: db}
}

func (r *runnerRepository) GetByID(ctx context.Context, id int64) (*runner.Runner, error) {
	var out runner.Runner
	if err := r.db.WithContext(ctx).First(&out, id).Error; err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *runnerRepository) GetByNodeID(ctx context.Context, nodeID string) (*runner.Runner, error) {
	var out runner.Runner
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).First(&out).Error; err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *runnerRepository) GetByNodeIDAndOrgID(ctx context.Context, nodeID string, orgID int64) (*runner.Runner, error) {
	var out runner.Runner
	if err := r.db.WithContext(ctx).
		Where("node_id = ? AND organization_id = ?", nodeID, orgID).
		First(&out).Error; err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *runnerRepository) ExistsByNodeIDAndOrg(ctx context.Context, orgID int64, nodeID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("organization_id = ? AND node_id = ?", orgID, nodeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *runnerRepository) Create(ctx context.Context, rn *runner.Runner) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateRunnerCluster(tx, rn); err != nil {
			return err
		}
		return tx.Create(rn).Error
	})
}

func (r *runnerRepository) UpdateFields(ctx context.Context, runnerID int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("id = ?", runnerID).
		Updates(updates).Error
}

func (r *runnerRepository) UpdateFieldsCAS(ctx context.Context, runnerID int64, casField string, casValue interface{}, updates map[string]interface{}) (int64, error) {
	result := r.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("id = ? AND "+casField+" = ?", runnerID, casValue).
		Updates(updates)
	return result.RowsAffected, result.Error
}

func (r *runnerRepository) Delete(ctx context.Context, runnerID int64) error {
	return r.db.WithContext(ctx).Delete(&runner.Runner{}, runnerID).Error
}

const visibilityWithGrantsFilter = "(visibility = 'organization' OR (visibility = 'private' AND registered_by_user_id = ?) OR CAST(id AS TEXT) IN (SELECT resource_id FROM resource_grants WHERE resource_type = ? AND user_id = ? AND organization_id = ?))"

func (r *runnerRepository) ListByOrg(ctx context.Context, orgID, userID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND "+visibilityWithGrantsFilter, orgID, userID, grant.TypeRunner, userID, orgID).
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

func (r *runnerRepository) ListForClusterStatus(ctx context.Context, orgID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

func (r *runnerRepository) ListAvailable(ctx context.Context, orgID, userID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND is_enabled = ? AND current_pods < max_concurrent_pods AND "+visibilityWithGrantsFilter,
			orgID, runner.RunnerStatusOnline, true, userID, grant.TypeRunner, userID, orgID).
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

func (r *runnerRepository) ListAvailableOrdered(ctx context.Context, orgID, userID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND is_enabled = ? AND current_pods < max_concurrent_pods AND "+visibilityWithGrantsFilter,
			orgID, runner.RunnerStatusOnline, true, userID, grant.TypeRunner, userID, orgID).
		Order("current_pods ASC").
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

func (r *runnerRepository) ListAvailableForAgent(ctx context.Context, orgID, userID int64, agentJSON string) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND is_enabled = ? AND current_pods < max_concurrent_pods AND available_agents @> ? AND "+visibilityWithGrantsFilter,
			orgID, runner.RunnerStatusOnline, true, agentJSON, userID, grant.TypeRunner, userID, orgID).
		Order("current_pods ASC").
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

func (r *runnerRepository) MarkOfflineRunners(ctx context.Context, threshold time.Time) error {
	return r.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("status = ? AND last_heartbeat < ?", runner.RunnerStatusOnline, threshold).
		Update("status", runner.RunnerStatusOffline).Error
}

func (r *runnerRepository) GetOrgSlug(ctx context.Context, orgID int64) (string, error) {
	var org struct{ Slug string }
	if err := r.db.WithContext(ctx).Table("organizations").
		Select("slug").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		if isNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return org.Slug, nil
}

func (r *runnerRepository) CountLoopsByRunner(ctx context.Context, runnerID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM workflows WHERE runner_id = ?", runnerID,
	).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
