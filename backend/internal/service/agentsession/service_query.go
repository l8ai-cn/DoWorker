package agentsession

import (
	"context"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
)

type ListOptions struct {
	Limit           int
	Project         string
	IncludeArchived bool
	PrincipalEmail  string
}

func (s *Service) ListForUser(ctx context.Context, orgID, userID int64, opts ListOptions) ([]domain.Session, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	q := s.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if opts.PrincipalEmail != "" {
		q = q.Where(
			"user_id = ? OR id IN (SELECT session_id FROM session_permissions WHERE user_id IN (?, '__public__'))",
			userID, opts.PrincipalEmail,
		)
	} else {
		q = q.Where("user_id = ?", userID)
	}
	if opts.Project != "" {
		q = q.Where("project = ?", opts.Project)
	}
	if !opts.IncludeArchived {
		q = q.Where("archived = ?", false)
	}
	var rows []domain.Session
	err := q.Order("updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (s *Service) ListProjects(ctx context.Context, orgID, userID int64) ([]string, error) {
	var names []string
	err := s.db.WithContext(ctx).Model(&domain.Session{}).
		Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL AND project IS NOT NULL AND project <> ''", orgID, userID).
		Distinct().
		Order("project ASC").
		Pluck("project", &names).Error
	return names, err
}

func (s *Service) GetActive(ctx context.Context, id string) (*domain.Session, error) {
	row, err := s.Get(ctx, id)
	if err != nil || row == nil || row.DeletedAt != nil {
		if err == nil && row != nil && row.DeletedAt != nil {
			return nil, ErrNotFound
		}
		return row, err
	}
	return row, nil
}
