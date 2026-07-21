package goalloop

import (
	"context"

	airesourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	airesourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
)

func (generator *DraftGenerator) resolveResource(
	ctx context.Context,
	scope DraftGenerationScope,
	resourceID int64,
) (*airesourceservice.ResolvedResource, error) {
	return generator.resources.ResolveExact(
		ctx,
		airesourceservice.Actor{UserID: scope.UserID},
		scope.OrganizationID,
		resourceID,
		airesourceservice.ResolutionRequirements{
			Modality:                airesourcedomain.ModalityChat,
			Capability:              airesourcedomain.CapabilityTextGeneration,
			AllowedProtocolAdapters: append([]string(nil), supportedDraftAdapters...),
		},
	)
}
