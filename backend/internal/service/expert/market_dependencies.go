package expert

import (
	"context"
	"fmt"
	"sort"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
)

func (s *Service) loadMarketSkills(
	ctx context.Context,
	ids []int64,
	cachedSlugs []string,
) ([]skilldom.Skill, error) {
	required := normalizeMarketSkillIDs(ids)
	skills, err := s.marketSkills.ListByIDs(ctx, required)
	if err != nil {
		return nil, err
	}
	found := make(map[int64]skilldom.Skill, len(skills))
	canonicalSlugs := make(map[string]struct{}, len(skills))
	missing := make([]string, 0)
	for _, skill := range skills {
		found[skill.ID] = skill
		canonicalSlugs[skill.Slug] = struct{}{}
		if skill.OrganizationID != nil || !skill.IsActive ||
			skill.ContentSha == "" || skill.StorageKey == "" {
			missing = append(missing, skill.Slug)
		}
	}
	unmatchedCached := make([]string, 0)
	for _, slug := range normalizeMarketStrings(cachedSlugs) {
		if _, ok := canonicalSlugs[slug]; !ok {
			unmatchedCached = append(unmatchedCached, slug)
		}
	}
	for _, id := range required {
		if _, ok := found[id]; ok {
			continue
		}
		if len(unmatchedCached) > 0 {
			missing = append(missing, unmatchedCached[0])
			unmatchedCached = unmatchedCached[1:]
		} else {
			missing = append(missing, fmt.Sprintf("skill-id-%d", id))
		}
	}
	if len(missing) > 0 {
		return nil, &MarketDependencyError{
			Missing: normalizeMarketStrings(missing),
		}
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Slug < skills[j].Slug
	})
	return skills, nil
}

func normalizeMarketSkillIDs(values []int64) []int64 {
	unique := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value > 0 {
			unique[value] = struct{}{}
		}
	}
	out := make([]int64, 0, len(unique))
	for value := range unique {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
