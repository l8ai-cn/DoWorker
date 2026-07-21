package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/tokenquota"
	"gorm.io/gorm"
)

var _ tokenquota.Repository = (*tokenQuotaRepo)(nil)

type tokenQuotaRepo struct{ db *gorm.DB }

func NewTokenQuotaRepository(db *gorm.DB) tokenquota.Repository {
	return &tokenQuotaRepo{db: db}
}

// Upsert matches on the (org, user, model) scope. The unique index uses
// COALESCE(user_id,0)/COALESCE(model,'') so ON CONFLICT can't target it via
// column names; find-then-write keeps the scope semantics explicit.
func (r *tokenQuotaRepo) Upsert(ctx context.Context, q *tokenquota.TokenQuota) error {
	db := r.db.WithContext(ctx)
	query := db.Model(&tokenquota.TokenQuota{}).Where("organization_id = ?", q.OrganizationID)
	if q.UserID == nil {
		query = query.Where("user_id IS NULL")
	} else {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Model == nil {
		query = query.Where("model IS NULL")
	} else {
		query = query.Where("model = ?", *q.Model)
	}

	var existing tokenquota.TokenQuota
	err := query.First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(q).Error
	}
	if err != nil {
		return err
	}
	q.ID = existing.ID
	return db.Model(&tokenquota.TokenQuota{}).Where("id = ?", existing.ID).
		Updates(map[string]interface{}{
			"limit_tokens": q.LimitTokens,
			"period":       q.Period,
		}).Error
}

func (r *tokenQuotaRepo) ListByOrg(ctx context.Context, orgID int64) ([]*tokenquota.TokenQuota, error) {
	var quotas []*tokenquota.TokenQuota
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("user_id NULLS FIRST, model NULLS FIRST").
		Find(&quotas).Error
	return quotas, err
}

func (r *tokenQuotaRepo) Delete(ctx context.Context, id, orgID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		Delete(&tokenquota.TokenQuota{}).Error
}
