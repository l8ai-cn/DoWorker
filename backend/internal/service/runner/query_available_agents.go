package runner

import (
	"context"
	"sort"
)

func (s *Service) ListAvailableAgentSlugs(ctx context.Context, orgID, userID int64) ([]string, error) {
	eligible, err := s.collectEligibleRunners(ctx, orgID, userID, "")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, active := range eligible {
		for _, slug := range active.Runner.AvailableAgents {
			seen[slug] = struct{}{}
		}
	}

	slugs := make([]string, 0, len(seen))
	for slug := range seen {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs, nil
}
