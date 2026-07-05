package agentsession

import (
	"context"
	"encoding/json"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
)

func (s *Service) UpdateAgentAndPod(ctx context.Context, id, agentSlug, podKey string) error {
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"agent_slug": agentSlug, "pod_key": podKey, "updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) SetMcpServers(ctx context.Context, id string, servers []domain.McpServer) error {
	raw, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"mcp_servers": raw, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) SetCodexGoal(ctx context.Context, id string, goal *domain.CodexGoal) error {
	var raw []byte
	var err error
	if goal == nil {
		raw = nil
	} else {
		raw, err = json.Marshal(goal)
		if err != nil {
			return err
		}
	}
	res := s.db.WithContext(ctx).Model(&struct{}{}).Table("agent_sessions").
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"codex_goal": raw, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) GetCodexGoal(ctx context.Context, id string) (*domain.CodexGoal, error) {
	row, err := s.GetActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(row.CodexGoal) == 0 || string(row.CodexGoal) == "null" {
		return nil, nil
	}
	var goal domain.CodexGoal
	if json.Unmarshal(row.CodexGoal, &goal) != nil {
		return nil, nil
	}
	return &goal, nil
}
