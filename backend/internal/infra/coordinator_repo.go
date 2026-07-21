package infra

import (
	"context"
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
	"gorm.io/gorm"
)

type coordinatorRepo struct {
	db *gorm.DB
}

func NewCoordinatorRepository(db *gorm.DB) coordinator.Repository {
	return &coordinatorRepo{db: db}
}

func (r *coordinatorRepo) CreateProject(ctx context.Context, project *coordinator.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *coordinatorRepo) GetProject(ctx context.Context, orgID, id int64) (*coordinator.Project, error) {
	var project coordinator.Project
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coordinator.ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *coordinatorRepo) GetProjectBySlug(ctx context.Context, orgID int64, slug string) (*coordinator.Project, error) {
	var project coordinator.Project
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coordinator.ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *coordinatorRepo) ListProjects(ctx context.Context, filter *coordinator.ProjectListFilter) ([]*coordinator.Project, error) {
	query := r.db.WithContext(ctx).Where("organization_id = ?", filter.OrganizationID)
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	var projects []*coordinator.Project
	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

func (r *coordinatorRepo) ListEnabledProjects(ctx context.Context) ([]*coordinator.Project, error) {
	var projects []*coordinator.Project
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&projects).Error
	return projects, err
}

func (r *coordinatorRepo) UpdateProject(ctx context.Context, orgID, id int64, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&coordinator.Project{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Updates(updates).Error
}

func (r *coordinatorRepo) DeleteProject(ctx context.Context, orgID, id int64) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&coordinator.Project{}).Error
}

func (r *coordinatorRepo) GetLinkByExternalID(ctx context.Context, orgID int64, platformType, externalID string) (*coordinator.TicketExternalLink, error) {
	var link coordinator.TicketExternalLink
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND platform_type = ? AND external_id = ?", orgID, platformType, externalID).
		First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coordinator.ErrNotFound
		}
		return nil, err
	}
	return &link, nil
}

func (r *coordinatorRepo) CreateLink(ctx context.Context, link *coordinator.TicketExternalLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *coordinatorRepo) CreateExecution(ctx context.Context, execution *coordinator.Execution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

func (r *coordinatorRepo) GetExecution(ctx context.Context, id int64) (*coordinator.Execution, error) {
	var execution coordinator.Execution
	if err := r.db.WithContext(ctx).First(&execution, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coordinator.ErrNotFound
		}
		return nil, err
	}
	return &execution, nil
}

func (r *coordinatorRepo) GetExecutionByPodKey(ctx context.Context, podKey string) (*coordinator.Execution, error) {
	var execution coordinator.Execution
	if err := r.db.WithContext(ctx).
		Where("pod_key = ?", podKey).
		Order("created_at DESC").
		First(&execution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coordinator.ErrNotFound
		}
		return nil, err
	}
	return &execution, nil
}

func (r *coordinatorRepo) ListExecutions(ctx context.Context, projectID int64, limit int) ([]*coordinator.Execution, error) {
	if limit <= 0 {
		limit = 50
	}
	var executions []*coordinator.Execution
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

func (r *coordinatorRepo) UpdateExecution(ctx context.Context, id int64, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&coordinator.Execution{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *coordinatorRepo) CountActiveExecutions(ctx context.Context, projectID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&coordinator.Execution{}).
		Where("project_id = ? AND status IN ?", projectID,
			[]string{coordinator.ExecutionStatusPending, coordinator.ExecutionStatusClaimed, coordinator.ExecutionStatusRunning}).
		Count(&count).Error
	return count, err
}

var _ coordinator.Repository = (*coordinatorRepo)(nil)
