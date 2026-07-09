package infra

import (
	"context"
	"errors"

	"gorm.io/gorm"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

// SkillCatalogRepository is the gorm-backed store for the unified skills
// catalog (git-backed rows, namespace am-skills).
type SkillCatalogRepository struct {
	db *gorm.DB
}

func NewSkillCatalogRepository(db *gorm.DB) *SkillCatalogRepository {
	return &SkillCatalogRepository{db: db}
}

func (r *SkillCatalogRepository) Create(ctx context.Context, s *skilldom.Skill) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SkillCatalogRepository) Update(ctx context.Context, s *skilldom.Skill) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *SkillCatalogRepository) Delete(ctx context.Context, orgID, id int64) error {
	res := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&skilldom.Skill{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return skilldom.ErrNotFound
	}
	return nil
}

func (r *SkillCatalogRepository) GetByID(ctx context.Context, orgID, id int64) (*skilldom.Skill, error) {
	var row skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&row).Error
	return skillRowOrNotFound(&row, err)
}

func (r *SkillCatalogRepository) GetAnyByID(ctx context.Context, id int64) (*skilldom.Skill, error) {
	var row skilldom.Skill
	err := r.db.WithContext(ctx).First(&row, id).Error
	return skillRowOrNotFound(&row, err)
}

func (r *SkillCatalogRepository) GetBySlug(ctx context.Context, orgID int64, slug string) (*skilldom.Skill, error) {
	var row skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&row).Error
	return skillRowOrNotFound(&row, err)
}

func (r *SkillCatalogRepository) FindByUpstream(ctx context.Context, orgID int64, upstreamURL, upstreamSubdir string) (*skilldom.Skill, error) {
	var row skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND upstream_url = ? AND upstream_subdir = ?", orgID, upstreamURL, upstreamSubdir).
		First(&row).Error
	return skillRowOrNotFound(&row, err)
}

func skillRowOrNotFound(row *skilldom.Skill, err error) (*skilldom.Skill, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, skilldom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (r *SkillCatalogRepository) SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&skilldom.Skill{}).
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

func (r *SkillCatalogRepository) List(ctx context.Context, orgID int64, limit, offset int) ([]skilldom.Skill, int64, error) {
	q := r.db.WithContext(ctx).Model(&skilldom.Skill{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	var rows []skilldom.Skill
	err := q.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *SkillCatalogRepository) ListCatalog(ctx context.Context, orgID int64, query, category string) ([]skilldom.Skill, error) {
	q := r.db.WithContext(ctx).Model(&skilldom.Skill{}).
		Where("is_active = ?", true).
		Where("organization_id IS NULL OR organization_id = ?", orgID)
	if query != "" {
		search := "%" + escapeLike(query) + "%"
		q = q.Where("slug ILIKE ? OR display_name ILIKE ? OR description ILIKE ?", search, search, search)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var rows []skilldom.Skill
	err := q.Order("display_name ASC, slug ASC").Find(&rows).Error
	return rows, err
}

var _ skilldom.Repository = (*SkillCatalogRepository)(nil)
