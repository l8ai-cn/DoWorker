package infra

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

type ExpertRepository struct {
	db *gorm.DB
}

func NewExpertRepository(db *gorm.DB) *ExpertRepository {
	return &ExpertRepository{db: db}
}

func (r *ExpertRepository) Create(ctx context.Context, expert *expertdom.Expert) error {
	return r.db.WithContext(ctx).Create(expert).Error
}

func (r *ExpertRepository) Update(ctx context.Context, expert *expertdom.Expert) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current expertdom.Expert
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"organization_id = ? AND id = ?",
				expert.OrganizationID,
				expert.ID,
			).
			First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return expertdom.ErrNotFound
		}
		if err != nil {
			return err
		}
		if current.Revision != expert.Revision {
			return expertdom.ErrConflict
		}
		expert.Revision++
		return tx.Save(expert).Error
	})
}

func (r *ExpertRepository) Delete(ctx context.Context, orgID, id int64) error {
	res := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&expertdom.Expert{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return expertdom.ErrNotFound
	}
	return nil
}

func (r *ExpertRepository) GetByID(ctx context.Context, orgID, id int64) (*expertdom.Expert, error) {
	var row expertdom.Expert
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertdom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ExpertRepository) GetBySlug(ctx context.Context, orgID int64, slug string) (*expertdom.Expert, error) {
	var row expertdom.Expert
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertdom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ExpertRepository) SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&expertdom.Expert{}).
		Where("organization_id = ? AND slug = ?", orgID, slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ExpertRepository) List(ctx context.Context, orgID int64, limit, offset int) ([]expertdom.Expert, int64, error) {
	q := r.db.WithContext(ctx).Model(&expertdom.Expert{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	var rows []expertdom.Expert
	err := q.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *ExpertRepository) RecordRun(ctx context.Context, orgID, id int64, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&expertdom.Expert{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Updates(map[string]interface{}{
			"run_count":   gorm.Expr("run_count + 1"),
			"last_run_at": at,
			"revision":    gorm.Expr("revision + 1"),
			"updated_at":  at,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return expertdom.ErrNotFound
	}
	return nil
}
