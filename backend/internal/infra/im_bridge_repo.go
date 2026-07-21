package infra

import (
	"context"
	"errors"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type imBridgeRepository struct{ db *gorm.DB }

func NewIMBridgeRepository(db *gorm.DB) domain.Repository {
	return &imBridgeRepository{db: db}
}

func (r *imBridgeRepository) ListConnections(ctx context.Context, orgID int64) ([]*domain.Connection, error) {
	var rows []*domain.Connection
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *imBridgeRepository) ListActiveByProvider(ctx context.Context, provider string) ([]*domain.Connection, error) {
	var rows []*domain.Connection
	err := r.db.WithContext(ctx).
		Where("provider = ? AND status = ?", provider, domain.StatusActive).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *imBridgeRepository) GetConnection(ctx context.Context, orgID, id int64) (*domain.Connection, error) {
	var row domain.Connection
	err := r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", orgID, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *imBridgeRepository) GetConnectionByToken(ctx context.Context, provider, token string) (*domain.Connection, error) {
	var row domain.Connection
	err := r.db.WithContext(ctx).Where("provider = ? AND webhook_token = ?", provider, token).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *imBridgeRepository) CreateConnection(ctx context.Context, conn *domain.Connection) error {
	return r.db.WithContext(ctx).Create(conn).Error
}

func (r *imBridgeRepository) UpdateConnection(ctx context.Context, conn *domain.Connection) error {
	return r.db.WithContext(ctx).Save(conn).Error
}

func (r *imBridgeRepository) DeleteConnection(ctx context.Context, orgID, id int64) error {
	return r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", orgID, id).Delete(&domain.Connection{}).Error
}

func (r *imBridgeRepository) GetThreadMapping(ctx context.Context, connectionID int64, externalThreadID string) (*domain.ThreadMapping, error) {
	var row domain.ThreadMapping
	err := r.db.WithContext(ctx).Where("connection_id = ? AND external_thread_id = ?", connectionID, externalThreadID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *imBridgeRepository) GetThreadMappingByChannel(ctx context.Context, connectionID, channelID int64) (*domain.ThreadMapping, error) {
	var row domain.ThreadMapping
	err := r.db.WithContext(ctx).Where("connection_id = ? AND channel_id = ?", connectionID, channelID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *imBridgeRepository) UpsertThreadMapping(ctx context.Context, mapping *domain.ThreadMapping) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "connection_id"}, {Name: "external_thread_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"channel_id", "context_token"}),
	}).Create(mapping).Error
}
