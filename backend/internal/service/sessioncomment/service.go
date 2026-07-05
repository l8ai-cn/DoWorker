package sessioncomment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/sessioncomment"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("comment not found")

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func NewID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cmt_" + hex.EncodeToString(b), nil
}

type CreateInput struct {
	SessionID, Path, Body string
	StartIndex, EndIndex  int
	AnchorContent         *string
	CreatedBy             *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Comment, error) {
	id, err := NewID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := &domain.Comment{
		ID: id, SessionID: in.SessionID, Path: in.Path,
		StartIndex: in.StartIndex, EndIndex: in.EndIndex, Body: in.Body,
		Status: "draft", AnchorContent: in.AnchorContent, CreatedBy: in.CreatedBy,
		CreatedAt: now, UpdatedAt: now,
	}
	return row, s.db.WithContext(ctx).Create(row).Error
}

func (s *Service) List(ctx context.Context, sessionID, path string) ([]domain.Comment, error) {
	q := s.db.WithContext(ctx).Where("session_id = ?", sessionID)
	if path != "" {
		q = q.Where("path = ?", path)
	}
	var rows []domain.Comment
	return rows, q.Order("created_at ASC").Find(&rows).Error
}

func (s *Service) Get(ctx context.Context, sessionID, id string) (*domain.Comment, error) {
	var row domain.Comment
	err := s.db.WithContext(ctx).Where("id = ? AND session_id = ?", id, sessionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &row, err
}

func (s *Service) Update(ctx context.Context, sessionID, id string, status, body *string) (*domain.Comment, error) {
	row, err := s.Get(ctx, sessionID, id)
	if err != nil {
		return nil, err
	}
	if status != nil && *status != "" {
		row.Status = *status
	}
	if body != nil {
		row.Body = *body
	}
	row.UpdatedAt = time.Now()
	return row, s.db.WithContext(ctx).Save(row).Error
}

func (s *Service) Delete(ctx context.Context, sessionID, id string) error {
	res := s.db.WithContext(ctx).Where("id = ? AND session_id = ?", id, sessionID).Delete(&domain.Comment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) MarkAddressed(ctx context.Context, sessionID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&domain.Comment{}).
		Where("session_id = ? AND id IN ?", sessionID, ids).
		Updates(map[string]any{"status": "addressed", "updated_at": time.Now()}).Error
}

func FormatSendMessage(comments []domain.Comment, instruction string) string {
	var b strings.Builder
	b.WriteString("Please address the following code review comments:\n\n")
	for _, c := range comments {
		fmt.Fprintf(&b, "File: %s (lines %d-%d)\n%s\n\n", c.Path, c.StartIndex+1, c.EndIndex, c.Body)
	}
	if strings.TrimSpace(instruction) != "" {
		b.WriteString(instruction)
	}
	return strings.TrimSpace(b.String())
}
