package runner

import (
	"context"
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

	value, ok := s.activeRunners.Load(runnerID)
	if !ok {
		return nil, ErrNoRunnerForAgent
	}
	lease, ok := value.(*ActiveRunner)
	if !ok {
		return nil, ErrNoRunnerForAgent
	}
	active := &ActiveRunner{
		Runner:   candidate,
		LastPing: lease.LastPing,
		PodCount: candidate.CurrentPods,
	}
	if !isRunnerAvailableForAgent(active, orgID, agentSlug) {
		return nil, ErrNoRunnerForAgent
	}
	return candidate, nil
}

func (s *Service) collectEligibleRunners(
	ctx context.Context,
	orgID, userID int64,
	agentSlug string,
) ([]*ActiveRunner, error) {
	grantedIDs, err := s.fetchGrantedRunnerIDs(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	var (
		result     []*ActiveRunner
		collectErr error
	)
	s.activeRunners.Range(func(key, value interface{}) bool {
		runnerID, ok := key.(int64)
		if !ok {
			return true
		}
		lease, ok := value.(*ActiveRunner)
		if !ok {
			return true
		}
		r, err := s.repo.GetByID(ctx, runnerID)
		if err != nil {
			collectErr = err
			return false
		}
		if r == nil {
			return true
		}
		active := &ActiveRunner{
			Runner:   r,
			LastPing: lease.LastPing,
			PodCount: r.CurrentPods,
		}
		if !isRunnerAvailableForAgent(active, orgID, agentSlug) {
			return true
		}
		if !isVisibleToUser(r, userID, grantedIDs) {
			return true
		}
		result = append(result, active)
		return true
	})
	return result, collectErr
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

func (s *Service) fetchGrantedRunnerIDs(ctx context.Context, orgID, userID int64) (map[int64]bool, error) {
	if s.grantQuerier == nil {
		return nil, nil
	}
	ids, err := s.grantQuerier.GetGrantedResourceIDs(ctx, grant.TypeRunner, userID, orgID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	m := make(map[int64]bool, len(ids))
	for _, idStr := range ids {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			m[id] = true
		}
	}
	return m, nil
}
