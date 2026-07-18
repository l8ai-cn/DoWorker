package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workflowRepo struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) workflow.WorkflowRepository {
	return &workflowRepo{db: db}
}

func (r *workflowRepo) Create(ctx context.Context, l *workflow.Workflow) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *workflowRepo) GetByID(ctx context.Context, id int64) (*workflow.Workflow, error) {
	var l workflow.Workflow
	if err := r.db.WithContext(ctx).First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workflow.ErrNotFound
		}
		return nil, err
	}
	return &l, nil
}

func (r *workflowRepo) GetBySlug(ctx context.Context, orgID int64, slug string) (*workflow.Workflow, error) {
	var l workflow.Workflow
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&l).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workflow.ErrNotFound
		}
		return nil, err
	}
	return &l, nil
}

func (r *workflowRepo) List(ctx context.Context, filter *workflow.ListWorkflowsFilter) ([]*workflow.Workflow, int64, error) {
	query := r.db.WithContext(ctx).Where("organization_id = ?", filter.OrganizationID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status != ?", workflow.StatusArchived)
	}
	if filter.ExecutionMode != "" {
		query = query.Where("execution_mode = ?", filter.ExecutionMode)
	}
	if filter.CronEnabled != nil {
		if *filter.CronEnabled {
			query = query.Where("cron_expression IS NOT NULL AND cron_expression != ''")
		} else {
			query = query.Where("cron_expression IS NULL OR cron_expression = ''")
		}
	}
	if filter.Query != "" {
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(filter.Query)
		q := "%" + escaped + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ? OR description ILIKE ?", q, q, q)
	}

	var total int64
	if err := query.Model(&workflow.Workflow{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}

	var workflows []*workflow.Workflow
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&workflows).Error; err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

func (r *workflowRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&workflow.Workflow{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete atomically deletes a workflow and its associated workflow_runs, but only if
// the workflow has no active (pending/running) runs.
func (r *workflowRepo) Delete(ctx context.Context, orgID int64, slug string) (int64, error) {
	var affected int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var l workflow.Workflow
		result := tx.
			Where("organization_id = ? AND slug = ?", orgID, slug).
			Where("NOT EXISTS (SELECT 1 FROM workflow_runs lr WHERE lr.workflow_id = workflows.id AND lr.status IN (?, ?) AND lr.finished_at IS NULL)",
				workflow.RunStatusPending, workflow.RunStatusRunning).
			First(&l)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				var count int64
				tx.Model(&workflow.Workflow{}).
					Where("organization_id = ? AND slug = ?", orgID, slug).
					Count(&count)
				if count > 0 {
					return workflow.ErrHasActiveRuns
				}
				return nil
			}
			return result.Error
		}

		if err := tx.Where("workflow_id = ?", l.ID).Delete(&workflow.WorkflowRun{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&l).Error; err != nil {
			return err
		}
		affected = 1
		return nil
	})
	return affected, err
}

func (r *workflowRepo) GetDueCronWorkflows(ctx context.Context, orgIDs []int64) ([]*workflow.Workflow, error) {
	var workflows []*workflow.Workflow
	query := r.db.WithContext(ctx).
		Where("status = ? AND cron_expression IS NOT NULL AND cron_expression != '' AND next_run_at <= ?",
			workflow.StatusEnabled, time.Now())
	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}
	err := query.Find(&workflows).Error
	return workflows, err
}

// ClaimCronWorkflow atomically claims a cron workflow with SKIP LOCKED and advances next_run_at.
func (r *workflowRepo) ClaimCronWorkflow(ctx context.Context, workflowID int64, nextRunAt *time.Time) (bool, error) {
	claimed := false

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var l workflow.Workflow
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("id = ? AND status = ? AND cron_expression IS NOT NULL AND cron_expression != '' AND next_run_at <= ?",
				workflowID, workflow.StatusEnabled, time.Now()).
			First(&l).Error
		if err != nil {
			return nil
		}

		if nextRunAt != nil {
			tx.Model(&l).Update("next_run_at", nextRunAt)
		} else {
			fallback := time.Now().Add(1 * time.Hour)
			tx.Model(&l).Update("next_run_at", fallback)
		}

		claimed = true
		return nil
	})

	return claimed, err
}

func (r *workflowRepo) FindWorkflowsNeedingNextRun(ctx context.Context, orgIDs []int64) ([]*workflow.Workflow, error) {
	var workflows []*workflow.Workflow
	query := r.db.WithContext(ctx).
		Where("status = ? AND cron_expression IS NOT NULL AND cron_expression != '' AND next_run_at IS NULL",
			workflow.StatusEnabled)
	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}
	err := query.Find(&workflows).Error
	return workflows, err
}

var _ workflow.WorkflowRepository = (*workflowRepo)(nil)
