package sessionpermission

import (
	"context"
	"errors"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionpermission"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("permission not found")

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, sessionID string) ([]domain.Grant, error) {
	var rows []domain.Grant
	return rows, s.db.WithContext(ctx).Where("session_id = ?", sessionID).Find(&rows).Error
}

func (s *Service) Upsert(ctx context.Context, sessionID, userID string, level int) (*domain.Grant, error) {
	row := &domain.Grant{SessionID: sessionID, UserID: userID, Level: level}
	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"level"}),
	}).Create(row).Error
	return row, err
}

func (s *Service) Delete(ctx context.Context, sessionID, userID string) error {
	res := s.db.WithContext(ctx).Where("session_id = ? AND user_id = ?", sessionID, userID).
		Delete(&domain.Grant{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) EffectiveLevel(ctx context.Context, sessionID string, principals ...string) (int, bool) {
	if len(principals) == 0 {
		return 0, false
	}
	var rows []domain.Grant
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND user_id IN ?", sessionID, principals).
		Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return 0, false
	}
	max := 0
	for _, r := range rows {
		if r.Level > max {
			max = r.Level
		}
	}
	return max, max > 0
}
