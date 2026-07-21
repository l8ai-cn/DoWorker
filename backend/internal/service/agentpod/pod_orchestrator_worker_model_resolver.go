package agentpod

import (
	"context"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	resourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
)

type ModelResourceResolver interface {
	ResolveExact(context.Context, resourcesvc.Actor, int64, int64, resourcesvc.ResolutionRequirements) (*resourcesvc.ResolvedResource, error)
}

func (o *PodOrchestrator) resolveWorkerModelResource(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	agentDef *agentDomain.Agent,
) (*resourcesvc.ResolvedResource, error) {
	if req.preResolvedDependencies != nil {
		model := req.preResolvedDependencies.Models.Primary
		if model == nil {
			return nil, nil
		}
		return o.artifactModelResource(ctx, req, *model)
	}
	requirements, needsResource := modelRequirementsForRequest(req, agentDef)
	if !needsResource {
		return nil, nil
	}
	if req.ModelResourceID == nil || *req.ModelResourceID <= 0 {
		return nil, ErrMissingModelResource
	}
	if o.modelResources == nil {
		return nil, ErrModelResourceResolverUnavailable
	}
	return o.modelResources.ResolveExact(
		ctx,
		resourcesvc.Actor{UserID: req.UserID},
		req.OrganizationID,
		*req.ModelResourceID,
		requirements,
	)
}

func modelRequirementsForRequest(
	req *OrchestrateCreatePodRequest,
	agentDef *agentDomain.Agent,
) (resourcesvc.ResolutionRequirements, bool) {
	if req != nil && req.preparedWorkerSpec != nil {
		binding := req.preparedWorkerSpec.Runtime.ModelBinding
		if binding.IsEmpty() {
			return resourcesvc.ResolutionRequirements{}, false
		}
		return chatRequirements(binding.ProtocolAdapter.String()), true
	}
	return modelResourceRequirements(req.AgentSlug, agentDef)
}
