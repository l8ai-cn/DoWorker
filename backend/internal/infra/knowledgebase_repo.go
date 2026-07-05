package infra

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"gorm.io/gorm"
)

type knowledgeBaseRepo struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepository(db *gorm.DB) knowledgebase.Repository {
	return &knowledgeBaseRepo{db: db}
}

func (r *knowledgeBaseRepo) Create(ctx context.Context, kb *knowledgebase.KnowledgeBase) error {
	return r.db.WithContext(ctx).Create(kb).Error
}

func (r *knowledgeBaseRepo) Get(ctx context.Context, orgID, id int64) (*knowledgebase.KnowledgeBase, error) {
	var kb knowledgebase.KnowledgeBase
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&kb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, knowledgebase.ErrNotFound
		}
		return nil, err
	}
	return &kb, nil
}

func (r *knowledgeBaseRepo) GetBySlug(ctx context.Context, orgID int64, slug string) (*knowledgebase.KnowledgeBase, error) {
	var kb knowledgebase.KnowledgeBase
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		First(&kb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, knowledgebase.ErrNotFound
		}
		return nil, err
	}
	return &kb, nil
}

func (r *knowledgeBaseRepo) List(ctx context.Context, filter *knowledgebase.ListFilter) ([]*knowledgebase.KnowledgeBase, error) {
	query := r.db.WithContext(ctx).Where("organization_id = ?", filter.OrganizationID)
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	var kbs []*knowledgebase.KnowledgeBase
	err := query.Order("created_at DESC").Find(&kbs).Error
	return kbs, err
}

func (r *knowledgeBaseRepo) ListExternal(ctx context.Context) ([]*knowledgebase.KnowledgeBase, error) {
	var kbs []*knowledgebase.KnowledgeBase
	err := r.db.WithContext(ctx).
		Where("source_type <> ?", knowledgebase.SourceTypeGit).
		Order("id").
		Find(&kbs).Error
	return kbs, err
}

func (r *knowledgeBaseRepo) ListBySlugs(ctx context.Context, orgID int64, slugs []string) ([]*knowledgebase.KnowledgeBase, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	var kbs []*knowledgebase.KnowledgeBase
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug IN ?", orgID, slugs).
		Find(&kbs).Error
	return kbs, err
}

func (r *knowledgeBaseRepo) Update(ctx context.Context, orgID, id int64, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&knowledgebase.KnowledgeBase{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Updates(updates).Error
}

func (r *knowledgeBaseRepo) Delete(ctx context.Context, orgID, id int64) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&knowledgebase.KnowledgeBase{}).Error
}

func (r *knowledgeBaseRepo) SlugExists(ctx context.Context, orgID int64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&knowledgebase.KnowledgeBase{}).
		Where("organization_id = ? AND slug = ?", orgID, slug).
		Count(&count).Error
	return count > 0, err
}

var _ knowledgebase.Repository = (*knowledgeBaseRepo)(nil)
