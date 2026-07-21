package runner

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

func (s *Service) GetByNodeID(ctx context.Context, nodeID string) (*runner.Runner, error) {
	r, err := s.repo.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrRunnerNotFound
	}
	return r, nil
}

func (s *Service) GetByNodeIDAndOrgID(ctx context.Context, nodeID string, orgID int64) (*runner.Runner, error) {
	r, err := s.repo.GetByNodeIDAndOrgID(ctx, nodeID, orgID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrRunnerNotFound
	}
	return r, nil
}

func (s *Service) UpdateLastSeen(ctx context.Context, runnerID int64) error {
	now := time.Now()
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"last_heartbeat": now,
		"status":         runner.RunnerStatusOnline,
	})
}

func (s *Service) GetRunner(ctx context.Context, runnerID int64) (*runner.Runner, error) {
	r, err := s.repo.GetByID(ctx, runnerID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrRunnerNotFound
	}
	return r, nil
}

func (s *Service) ListRunners(ctx context.Context, orgID int64, userID int64) ([]*runner.Runner, error) {
	return s.repo.ListByOrg(ctx, orgID, userID)
}

func (s *Service) ListAvailableRunners(ctx context.Context, orgID int64, userID int64) ([]*runner.Runner, error) {
	return s.repo.ListAvailable(ctx, orgID, userID)
}

func (s *Service) SelectAvailableRunner(ctx context.Context, orgID int64, userID int64) (*runner.Runner, error) {
	cachedRunners, err := s.collectEligibleRunners(ctx, orgID, userID, "")
	if err != nil {
		return nil, err
	}

	if len(cachedRunners) > 0 {
		sort.Slice(cachedRunners, func(i, j int) bool {
			return cachedRunners[i].PodCount < cachedRunners[j].PodCount
		})
		return cachedRunners[0].Runner, nil
	}

	slog.WarnContext(ctx, "no connected runner available", "org_id", orgID, "user_id", userID)
	return nil, ErrRunnerOffline
}

func (s *Service) SelectAvailableRunnerForAgent(ctx context.Context, orgID int64, userID int64, agentSlug string) (*runner.Runner, error) {
	cachedRunners, err := s.collectEligibleRunners(ctx, orgID, userID, agentSlug)
	if err != nil {
		return nil, err
	}

	if len(cachedRunners) > 0 {
		sort.Slice(cachedRunners, func(i, j int) bool {
			return cachedRunners[i].PodCount < cachedRunners[j].PodCount
		})
		return cachedRunners[0].Runner, nil
	}

	slog.WarnContext(ctx, "no connected runner available for agent", "org_id", orgID, "agent_slug", agentSlug)
	return nil, ErrNoRunnerForAgent
}
