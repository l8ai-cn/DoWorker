package runner

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (s *Service) MostRecentRunnerForAgent(ctx context.Context, orgID, userID int64, agentSlug string) (*runnerDomain.Runner, error) {
	runners, err := s.repo.ListByOrg(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	var best *runnerDomain.Runner
	var bestHeartbeat time.Time
	for _, r := range runners {
		if !r.IsEnabled || !r.SupportsAgent(agentSlug) {
			continue
		}
		hb := r.UpdatedAt
		if r.LastHeartbeat != nil && r.LastHeartbeat.After(hb) {
			hb = *r.LastHeartbeat
		}
		if best == nil || hb.After(bestHeartbeat) {
			copy := *r
			best = &copy
			bestHeartbeat = hb
		}
	}
	if best == nil {
		return nil, ErrNoRunnerForAgent
	}
	return best, nil
}

func (s *Service) FirstAvailableAgentSlug(ctx context.Context, orgID, userID int64) (string, error) {
	runners, err := s.repo.ListByOrg(ctx, orgID, userID)
	if err != nil {
		return "", err
	}
	slugs := map[string]struct{}{}
	for _, r := range runners {
		if !r.IsEnabled {
			continue
		}
		for _, slug := range r.AvailableAgents {
			slugs[slug] = struct{}{}
		}
	}
	if len(slugs) == 0 {
		return "", ErrNoRunnerForAgent
	}
	list := make([]string, 0, len(slugs))
	for slug := range slugs {
		list = append(list, slug)
	}
	sort.Strings(list)
	return list[0], nil
}

func AgentSlugJSON(slug string) (string, error) {
	b, err := json.Marshal([]string{slug})
	return string(b), err
}
