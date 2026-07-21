package agentsession

import (
	"context"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
)

func (s *Service) ListChildren(ctx context.Context, parentID string) ([]domain.Session, error) {
	var rows []domain.Session
	err := s.db.WithContext(ctx).
		Where("parent_session_id = ? AND deleted_at IS NULL", parentID).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}
