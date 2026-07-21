package infra

import (
	"context"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
)

func (r *SkillCatalogRepository) GetPlatformBySlug(
	ctx context.Context,
	slug string,
) (*skilldom.Skill, error) {
	var row skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("organization_id IS NULL AND slug = ?", slug).
		First(&row).Error
	return skillRowOrNotFound(&row, err)
}

func (r *SkillCatalogRepository) ListActivePlatformBySlugs(
	ctx context.Context,
	slugs []string,
) ([]skilldom.Skill, error) {
	if len(slugs) == 0 {
		return []skilldom.Skill{}, nil
	}
	var rows []skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("organization_id IS NULL").
		Where("is_active = ?", true).
		Where("slug IN ?", slugs).
		Order("slug ASC").
		Find(&rows).Error
	return rows, err
}

func (r *SkillCatalogRepository) ListByIDs(
	ctx context.Context,
	ids []int64,
) ([]skilldom.Skill, error) {
	if len(ids) == 0 {
		return []skilldom.Skill{}, nil
	}
	var rows []skilldom.Skill
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}
