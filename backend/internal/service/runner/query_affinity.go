package runner

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (s *Service) SelectRunnerWithAffinity(
	ctx context.Context,
	orgID int64, userID int64, agentSlug string,
	hints *runner.AffinityHints,
	repoHistory map[int64]int,
) (*runner.Runner, error) {
	if hints == nil {
		return s.SelectAvailableRunnerForAgent(ctx, orgID, userID, agentSlug)
	}

	cachedRunners, err := s.collectEligibleRunners(ctx, orgID, userID, agentSlug)
	if err != nil {
		return nil, err
	}

	if len(cachedRunners) > 0 {
		return s.selectWithScoring(cachedRunners, userID, hints, repoHistory)
	}

	slog.Warn("no connected runner available for agent", "org_id", orgID, "agent_slug", agentSlug)
	return nil, ErrNoRunnerForAgent
}

func (s *Service) selectWithScoring(
	candidates []*ActiveRunner,
	userID int64,
	hints *runner.AffinityHints,
	repoHistory map[int64]int,
) (*runner.Runner, error) {
	ranked := ScoreRunners(candidates, userID, hints, repoHistory, runner.DefaultAffinityWeights())
	if len(ranked) == 0 {
		return nil, ErrNoRunnerForAgent
	}
	slog.Info("runner selected with affinity",
		"runner_id", ranked[0].Runner.ID,
		"score_count", len(ranked),
		"has_repo_hint", hints.RepositoryID != nil,
		"has_tags", len(hints.Tags) > 0,
	)
	return ranked[0].Runner, nil
}
