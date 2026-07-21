package envbundle

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
)

func (s *Service) GetEffectiveByIDs(
	ctx context.Context,
	userID, orgID int64,
	agentSlug string,
	ids []int64,
) ([]*EffectiveBundle, error) {
	bundles := make([]*EffectiveBundle, 0, len(ids))
	seenIDs := make(map[int64]struct{}, len(ids))
	seenNames := make(map[string]int64, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("env bundle id must be positive")
		}
		if _, exists := seenIDs[id]; exists {
			return nil, fmt.Errorf("env bundle id %d is duplicated", id)
		}
		seenIDs[id] = struct{}{}
		bundle, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("load env bundle %d: %w", id, err)
		}
		if !bundleAvailableToWorker(bundle, userID, orgID, agentSlug) {
			return nil, fmt.Errorf("%w: env bundle %d", ErrNotFound, id)
		}
		if existingID, exists := seenNames[bundle.Name]; exists {
			return nil, fmt.Errorf(
				"env bundles %d and %d share name %q",
				existingID,
				id,
				bundle.Name,
			)
		}
		data, err := s.decryptData(bundle.Kind, bundle.Data)
		if err != nil {
			return nil, fmt.Errorf("decrypt env bundle %d: %w", id, err)
		}
		seenNames[bundle.Name] = id
		bundles = append(bundles, &EffectiveBundle{
			ID:         bundle.ID,
			Name:       bundle.Name,
			Kind:       bundle.Kind,
			OwnerScope: bundle.OwnerScope,
			AgentSlug:  bundle.AgentSlug,
			Data:       data,
		})
	}
	return bundles, nil
}

func bundleAvailableToWorker(
	bundle *envbundle.EnvBundle,
	userID, orgID int64,
	agentSlug string,
) bool {
	if bundle == nil || !bundle.IsActive {
		return false
	}
	switch bundle.OwnerScope {
	case envbundle.OwnerScopeUser:
		if bundle.OwnerID != userID {
			return false
		}
	case envbundle.OwnerScopeOrg:
		if bundle.OwnerID != orgID {
			return false
		}
	default:
		return false
	}
	return bundle.AgentSlug == nil || *bundle.AgentSlug == agentSlug
}
