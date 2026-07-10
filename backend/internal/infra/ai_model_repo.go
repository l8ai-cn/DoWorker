package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"gorm.io/gorm"
)

var _ aimodel.Repository = (*aiModelRepo)(nil)

type aiModelRepo struct{ db *gorm.DB }

func NewAIModelRepository(db *gorm.DB) aimodel.Repository {
	return &aiModelRepo{db: db}
}

func (r *aiModelRepo) GetByID(ctx context.Context, id int64) (*aimodel.AIModel, error) {
	var m aimodel.AIModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiModelRepo) GetVisibleByID(ctx context.Context, id, userID, orgID int64) (*aimodel.AIModel, error) {
	var m aimodel.AIModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_enabled = ? AND (organization_id = ? OR user_id = ?)", id, true, orgID, userID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiModelRepo) Create(ctx context.Context, m *aimodel.AIModel) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *aiModelRepo) Save(ctx context.Context, m *aimodel.AIModel) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *aiModelRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&aimodel.AIModel{}, id).Error
}

func (r *aiModelRepo) ListVisible(ctx context.Context, userID, orgID int64) ([]*aimodel.AIModel, error) {
	var models []*aimodel.AIModel
	err := r.db.WithContext(ctx).
		Where("is_enabled = ? AND (organization_id = ? OR user_id = ?)", true, orgID, userID).
		Order("is_default DESC, provider_type, name").
		Find(&models).Error
	return models, err
}

func (r *aiModelRepo) DefaultVisible(ctx context.Context, userID, orgID int64) (*aimodel.AIModel, error) {
	var m aimodel.AIModel
	// user-private default wins over org default.
	err := r.db.WithContext(ctx).
		Where("is_enabled = ? AND is_default = ? AND (organization_id = ? OR user_id = ?)", true, true, orgID, userID).
		Order("user_id DESC NULLS LAST").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiModelRepo) ClearDefaults(ctx context.Context, userID, orgID int64) error {
	q := r.db.WithContext(ctx).Model(&aimodel.AIModel{})
	if orgID > 0 && userID == 0 {
		q = q.Where("organization_id = ? AND user_id IS NULL", orgID)
	} else {
		q = q.Where("user_id = ?", userID)
	}
	return q.Update("is_default", false).Error
}

func (r *aiModelRepo) CountOrg(ctx context.Context, orgID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&aimodel.AIModel{}).
		Where("organization_id = ?", orgID).Count(&n).Error
	return n, err
}

func (r *aiModelRepo) FirstVisibleByProvider(
	ctx context.Context, userID, orgID int64, providerType string,
) (*aimodel.AIModel, error) {
	var m aimodel.AIModel
	err := r.db.WithContext(ctx).
		Where(
			"is_enabled = ? AND provider_type = ? AND (organization_id = ? OR user_id = ?)",
			true, providerType, orgID, userID,
		).
		Order("is_default DESC, id").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
