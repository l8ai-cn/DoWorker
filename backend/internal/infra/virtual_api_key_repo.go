package infra

import (
	"context"
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/virtualkey"
	"gorm.io/gorm"
)

var _ virtualkey.Repository = (*virtualAPIKeyRepo)(nil)

type virtualAPIKeyRepo struct{ db *gorm.DB }

func NewVirtualAPIKeyRepository(db *gorm.DB) virtualkey.Repository {
	return &virtualAPIKeyRepo{db: db}
}

func (r *virtualAPIKeyRepo) Create(ctx context.Context, k *virtualkey.VirtualAPIKey) error {
	return r.db.WithContext(ctx).Create(k).Error
}

func (r *virtualAPIKeyRepo) GetByID(ctx context.Context, id int64) (*virtualkey.VirtualAPIKey, error) {
	var k virtualkey.VirtualAPIKey
	err := r.db.WithContext(ctx).First(&k, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *virtualAPIKeyRepo) GetByIDForScope(
	ctx context.Context, id, orgID, userID int64,
) (*virtualkey.VirtualAPIKey, error) {
	var k virtualkey.VirtualAPIKey
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND user_id = ?", id, orgID, userID).
		First(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *virtualAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*virtualkey.VirtualAPIKey, error) {
	var k virtualkey.VirtualAPIKey
	err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *virtualAPIKeyRepo) ListByScope(ctx context.Context, orgID, userID int64) ([]*virtualkey.VirtualAPIKey, error) {
	var keys []*virtualkey.VirtualAPIKey
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

func (r *virtualAPIKeyRepo) UpdateStatusForScope(
	ctx context.Context,
	id, orgID, userID int64,
	status string,
) (bool, error) {
	result := r.db.WithContext(ctx).Model(&virtualkey.VirtualAPIKey{}).
		Where("id = ? AND organization_id = ? AND user_id = ?", id, orgID, userID).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()})
	return result.RowsAffected > 0, result.Error
}

func (r *virtualAPIKeyRepo) TouchLastUsed(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&virtualkey.VirtualAPIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}
