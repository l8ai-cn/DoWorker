package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	"gorm.io/gorm"
)

type goalLoopRepo struct {
	db *gorm.DB
}

func NewGoalLoopRepository(db *gorm.DB) goalloop.Repository {
	return &goalLoopRepo{db: db}
}

func (r *goalLoopRepo) Create(ctx context.Context, loop *goalloop.GoalLoop) error {
	return r.db.WithContext(ctx).Create(loop).Error
}

func (r *goalLoopRepo) GetBySlug(ctx context.Context, orgID int64, slug string) (*goalloop.GoalLoop, error) {
	var loop goalloop.GoalLoop
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND slug = ?", orgID, slug).First(&loop).Error; err != nil {
		return nil, mapGoalLoopError(err)
	}
	return &loop, nil
}

func (r *goalLoopRepo) GetByPodKey(ctx context.Context, podKey string) (*goalloop.GoalLoop, error) {
	var loop goalloop.GoalLoop
	if err := r.db.WithContext(ctx).Where("pod_key = ?", podKey).First(&loop).Error; err != nil {
		return nil, mapGoalLoopError(err)
	}
	return &loop, nil
}

func (r *goalLoopRepo) GetByAutopilotControllerKey(ctx context.Context, autopilotKey string) (*goalloop.GoalLoop, error) {
	var loop goalloop.GoalLoop
	if err := r.db.WithContext(ctx).
		Where("autopilot_controller_key = ?", autopilotKey).
		First(&loop).Error; err != nil {
		return nil, mapGoalLoopError(err)
	}
	return &loop, nil
}

func (r *goalLoopRepo) GetByVerificationRequestID(ctx context.Context, requestID string) (*goalloop.GoalLoop, error) {
	var loop goalloop.GoalLoop
	if err := r.db.WithContext(ctx).Where("verification_request_id = ?", requestID).First(&loop).Error; err != nil {
		return nil, mapGoalLoopError(err)
	}
	return &loop, nil
}

func (r *goalLoopRepo) ListTimedOut(ctx context.Context, now time.Time) ([]*goalloop.GoalLoop, error) {
	var loops []*goalloop.GoalLoop
	err := r.db.WithContext(ctx).
		Where(
			"(status IN ? AND started_at IS NOT NULL AND "+
				"started_at + (timeout_minutes * INTERVAL '1 minute') <= ?) OR "+
				"(pod_key IS NOT NULL AND verification_error LIKE ?)",
			[]string{goalloop.StatusActive, goalloop.StatusVerifying},
			now,
			goalloop.PendingPodCleanupErrorPrefix+"%",
		).
		Find(&loops).Error
	return loops, err
}

func (r *goalLoopRepo) List(ctx context.Context, filter goalloop.ListFilter) ([]*goalloop.GoalLoop, int64, error) {
	query := r.db.WithContext(ctx).Where("organization_id = ?", filter.OrganizationID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Query != "" {
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(filter.Query)
		like := "%" + escaped + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ? OR objective ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Model(&goalloop.GoalLoop{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	var loops []*goalloop.GoalLoop
	if err := query.Order("created_at DESC").Limit(limit).Offset(filter.Offset).Find(&loops).Error; err != nil {
		return nil, 0, err
	}
	return loops, total, nil
}

func (r *goalLoopRepo) ExistsSlug(ctx context.Context, orgID int64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&goalloop.GoalLoop{}).
		Where("organization_id = ? AND slug = ?", orgID, slug).Count(&count).Error
	return count > 0, err
}

func (r *goalLoopRepo) Update(ctx context.Context, id int64, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&goalloop.GoalLoop{}).Where("id = ?", id).Updates(updates).Error
}

func (r *goalLoopRepo) TransitionStatus(
	ctx context.Context,
	id int64,
	from []string,
	updates map[string]any,
) (bool, error) {
	updates["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).
		Model(&goalloop.GoalLoop{}).
		Where("id = ? AND status IN ?", id, from).
		Updates(updates)
	return result.RowsAffected == 1, result.Error
}

func mapGoalLoopError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return goalloop.ErrNotFound
	}
	return err
}

var _ goalloop.Repository = (*goalLoopRepo)(nil)
