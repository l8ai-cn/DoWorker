package agentsession

import (
	"context"
	"time"
)

func (s *Service) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&struct {
		ID string `gorm:"primaryKey"`
	}{}).
		Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdateTitle(ctx context.Context, id string, title *string) error {
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"title": title, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdateArchived(ctx context.Context, id string, archived bool) error {
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"archived": archived, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdateProject(ctx context.Context, id string, project *string) error {
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"project": project, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
