package infra

import (
	"context"
	"errors"

	"gorm.io/gorm"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

// AuthoredSkillRepository is the gorm-backed store for platform-authored,
// git-backed skills (namespace am-skills). Mirrors ExpertRepository.
type AuthoredSkillRepository struct {
	db *gorm.DB
}

func NewAuthoredSkillRepository(db *gorm.DB) *AuthoredSkillRepository {
	return &AuthoredSkillRepository{db: db}
}

func (r *AuthoredSkillRepository) Create(ctx context.Context, s *skilldom.AuthoredSkill) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *AuthoredSkillRepository) Update(ctx context.Context, s *skilldom.AuthoredSkill) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *AuthoredSkillRepository) Delete(ctx context.Context, orgID, id int64) error {
	res := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&skilldom.AuthoredSkill{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return skilldom.ErrNotFound
	}
	return nil
}

func (r *AuthoredSkillRepository) GetByID(ctx context.Context, orgID, id int64) (*skilldom.AuthoredSkill, error) {
	var row skilldom.AuthoredSkill
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, skilldom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AuthoredSkillRepository) GetBySlug(ctx context.Context, orgID int64, slug string) (*skilldom.AuthoredSkill, error) {
	var row skilldom.AuthoredSkill
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, skilldom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AuthoredSkillRepository) SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&skilldom.AuthoredSkill{}).
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

func (r *AuthoredSkillRepository) List(ctx context.Context, orgID int64, limit, offset int) ([]skilldom.AuthoredSkill, int64, error) {
	q := r.db.WithContext(ctx).Model(&skilldom.AuthoredSkill{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	var rows []skilldom.AuthoredSkill
	err := q.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}
