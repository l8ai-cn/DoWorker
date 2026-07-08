package envbundle

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
)

// PrimaryCredentialBundleName returns the bundle name to auto-mount for agentSlug
// at session create. User-scoped primaries beat org-scoped; agent-specific rows
// beat universal (agent_slug IS NULL).
func (s *Service) PrimaryCredentialBundleName(ctx context.Context, userID, orgID int64, agentSlug string) (string, error) {
	bundles, err := s.repo.ListEffectiveForUser(ctx, userID, orgID, agentSlug)
	if err != nil {
		return "", err
	}
	var best *envbundle.EnvBundle
	score := -1
	for _, b := range bundles {
		if !b.IsActive || b.Kind != envbundle.KindCredential || !b.KindPrimary {
			continue
		}
		s := primaryScore(b, agentSlug)
		if s > score {
			score = s
			best = b
		}
	}
	if best == nil {
		return "", nil
	}
	return best.Name, nil
}

func primaryScore(b *envbundle.EnvBundle, agentSlug string) int {
	score := 0
	if b.OwnerScope == envbundle.OwnerScopeUser {
		score += 10
	}
	if b.AgentSlug != nil && *b.AgentSlug == agentSlug {
		score += 5
	}
	return score
}
