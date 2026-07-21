package agentsession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("agent session not found")

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func NewID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("session id: %w", err)
	}
	return "conv_" + hex.EncodeToString(b), nil
}

func (s *Service) Create(ctx context.Context, row *domain.Session) error {
	return s.db.WithContext(ctx).Create(row).Error
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Session, error) {
	var row domain.Session
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) ListByUser(ctx context.Context, orgID, userID int64, limit int) ([]domain.Session, error) {
	return s.ListForUser(ctx, orgID, userID, ListOptions{Limit: limit, IncludeArchived: true})
}

func (s *Service) GetByPodKey(ctx context.Context, podKey string) (*domain.Session, error) {
	var row domain.Session
	err := s.db.WithContext(ctx).
		Where("pod_key = ? AND deleted_at IS NULL", podKey).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) UpdateRunner(ctx context.Context, id, runnerNodeID string) error {
	res := s.db.WithContext(ctx).Model(&domain.Session{}).
		Where("id = ?", id).
		Update("runner_node_id", runnerNodeID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) TouchUpdatedAt(ctx context.Context, id string) error {
	res := s.db.WithContext(ctx).Model(&domain.Session{}).
		Where("id = ?", id).
		Update("updated_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdatePodKey(ctx context.Context, id, podKey string) error {
	res := s.db.WithContext(ctx).Model(&domain.Session{}).
		Where("id = ?", id).
		Update("pod_key", podKey)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
