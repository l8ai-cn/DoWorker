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
	if candidate == nil || !candidate.IsEnabled || !candidate.SupportsAgent(agentSlug) {
		return nil, ErrNoRunnerForAgent
	}
	if allowUnavailable {
		return candidate, nil
	}

	for _, active := range s.collectEligibleRunners(ctx, orgID, userID, agentSlug) {
		if active.Runner.ID == runnerID {
			return candidate, nil
		}
	}
	return nil, ErrNoRunnerForAgent
}

func (s *Service) collectEligibleRunners(ctx context.Context, orgID, userID int64, agentSlug string) []*ActiveRunner {
	grantedIDs := s.fetchGrantedRunnerIDs(ctx, orgID, userID)

	var result []*ActiveRunner
	s.activeRunners.Range(func(key, value interface{}) bool {
		ar, ok := value.(*ActiveRunner)
		if !ok || ar.Runner == nil {
			return true
		}
		r := ar.Runner
		if r.OrganizationID != orgID ||
			r.Status != runnerDomain.RunnerStatusOnline ||
			!r.IsEnabled ||
			ar.PodCount >= r.MaxConcurrentPods ||
			time.Since(ar.LastPing) >= 90*time.Second {
			return true
		}
		if agentSlug != "" && !r.SupportsAgent(agentSlug) {
			return true
		}
		if !isVisibleToUser(r, userID, grantedIDs) {
			return true
		}
		result = append(result, ar)
		return true
	})
	return result
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
