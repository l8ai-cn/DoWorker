package conversationitem

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func NewItemID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("item id: %w", err)
	}
	return "item_" + hex.EncodeToString(b), nil
}

func NewResponseID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("response id: %w", err)
	}
	return "resp_" + hex.EncodeToString(b), nil
}

func (s *Service) NextPosition(ctx context.Context, sessionID string) (int64, error) {
	var maxPos *int64
	err := s.db.WithContext(ctx).Model(&domain.Item{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return 1, nil
	}
	return *maxPos + 1, nil
}

func (s *Service) Append(ctx context.Context, row *domain.Item) error {
	return s.db.WithContext(ctx).Create(row).Error
}

type Page struct {
	Items   []domain.Item
	HasMore bool
}

func (s *Service) ListPage(ctx context.Context, sessionID string, limit int, afterID string, desc bool) (Page, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := s.db.WithContext(ctx).Where("session_id = ?", sessionID)
	if afterID != "" {
		var cursor domain.Item
		if err := s.db.WithContext(ctx).Where("id = ? AND session_id = ?", afterID, sessionID).First(&cursor).Error; err != nil {
			return Page{}, err
		}
		if desc {
			q = q.Where("position < ?", cursor.Position)
		} else {
			q = q.Where("position > ?", cursor.Position)
		}
	}
	order := "position ASC"
	if desc {
		order = "position DESC"
	}
	var rows []domain.Item
	if err := q.Order(order).Limit(limit + 1).Find(&rows).Error; err != nil {
		return Page{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return Page{Items: rows, HasMore: hasMore}, nil
}
