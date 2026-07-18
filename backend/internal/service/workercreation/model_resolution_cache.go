package workercreation

import resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"

func (resolver *modelResolver) resolvedModel(
	resourceID int64,
) (*resourceservice.ResolvedResource, bool) {
	if resolver == nil {
		return nil, false
	}
	resolved, ok := resolver.resolved[resourceID]
	return resolved, ok
}
