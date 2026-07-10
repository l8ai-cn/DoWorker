package runner

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/grant"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (s *Service) ResolveRunnerForCreate(
	ctx context.Context,
	runnerID, orgID, userID int64,
	agentSlug string,
	allowUnavailable bool,
) (*runnerDomain.Runner, error) {
	runners, err := s.repo.ListByOrg(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	var candidate *runnerDomain.Runner
	for _, r := range runners {
		if r.ID == runnerID {
			candidate = r
			break
		}
	}
	if candidate == nil || !candidate.IsEnabled {
		return nil, ErrNoRunnerForAgent
	}

	value, ok := s.activeRunners.Load(runnerID)
	var active *ActiveRunner
	if ok {
		active, _ = value.(*ActiveRunner)
	}
	if !runnerSupportsAgent(candidate, active, agentSlug) {
		return nil, ErrNoRunnerForAgent
	}
	if allowUnavailable {
		return candidate, nil
	}
	if active != nil {
		if !isRunnerAvailableForAgent(withAgentFallback(active, candidate), orgID, agentSlug) {
			return nil, ErrNoRunnerForAgent
		}
		return candidate, nil
	}
	// activeRunners is only populated after UpdateLastSeen/MarkConnected; until
	// then mirror SelectRunnerWithAffinity's DB fallback so explicit placement
	// works the same as auto-select.
	if candidate.OrganizationID != orgID ||
		candidate.Status != runnerDomain.RunnerStatusOnline ||
		candidate.CurrentPods >= candidate.MaxConcurrentPods {
		return nil, ErrNoRunnerForAgent
	}
	return candidate, nil
}

func runnerSupportsAgent(candidate *runnerDomain.Runner, active *ActiveRunner, agentSlug string) bool {
	if agentSlug == "" {
		return true
	}
	if candidate != nil && candidate.SupportsAgent(agentSlug) {
		return true
	}
	return active != nil && active.Runner != nil && active.Runner.SupportsAgent(agentSlug)
}

func withAgentFallback(active *ActiveRunner, candidate *runnerDomain.Runner) *ActiveRunner {
	if active == nil || active.Runner == nil || candidate == nil {
		return active
	}
	if len(active.Runner.AvailableAgents) > 0 || len(candidate.AvailableAgents) == 0 {
		return active
	}
	patched := *active.Runner
	patched.AvailableAgents = candidate.AvailableAgents
	return &ActiveRunner{Runner: &patched, LastPing: active.LastPing, PodCount: active.PodCount}
}

func (s *Service) collectEligibleRunners(ctx context.Context, orgID, userID int64, agentSlug string) []*ActiveRunner {
	grantedIDs := s.fetchGrantedRunnerIDs(ctx, orgID, userID)

	var result []*ActiveRunner
	s.activeRunners.Range(func(key, value interface{}) bool {
		ar, ok := value.(*ActiveRunner)
		if !ok || !isRunnerAvailableForAgent(ar, orgID, agentSlug) {
			return true
		}
		r := ar.Runner
		if !isVisibleToUser(r, userID, grantedIDs) {
			return true
		}
		result = append(result, ar)
		return true
	})
	return result
}

func isRunnerAvailableForAgent(ar *ActiveRunner, orgID int64, agentSlug string) bool {
	if ar == nil || ar.Runner == nil {
		return false
	}
	r := ar.Runner
	if r.OrganizationID != orgID ||
		r.Status != runnerDomain.RunnerStatusOnline ||
		!r.IsEnabled ||
		ar.PodCount >= r.MaxConcurrentPods ||
		time.Since(ar.LastPing) >= 90*time.Second {
		return false
	}
	return agentSlug == "" || r.SupportsAgent(agentSlug)
}

func isVisibleToUser(r *runnerDomain.Runner, userID int64, grantedIDs map[int64]bool) bool {
	if r.Visibility == runnerDomain.VisibilityOrganization {
		return true
	}
	if r.RegisteredByUserID != nil && *r.RegisteredByUserID == userID {
		return true
	}
	return grantedIDs[r.ID]
}

func (s *Service) fetchGrantedRunnerIDs(ctx context.Context, orgID, userID int64) map[int64]bool {
	if s.grantQuerier == nil {
		return nil
	}
	ids, err := s.grantQuerier.GetGrantedResourceIDs(ctx, grant.TypeRunner, userID, orgID)
	if err != nil {
		slog.Warn("failed to fetch runner grants for cache filter", "user_id", userID, "error", err)
		return nil
	}
	if len(ids) == 0 {
		return nil
	}
	m := make(map[int64]bool, len(ids))
	for _, idStr := range ids {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			m[id] = true
		}
	}
	return m
}
