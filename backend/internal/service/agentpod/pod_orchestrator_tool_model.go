package agentpod

import (
	"context"
	"strings"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

func (o *PodOrchestrator) applyWorkerToolModels(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) error {
	if req == nil || req.preparedWorkerSpec == nil ||
		len(req.preparedWorkerSpec.Runtime.ToolModelBindings) == 0 {
		return nil
	}
	if o.modelResources == nil {
		return ErrModelResourceResolverUnavailable
	}
	environment := map[string]string{}
	for _, binding := range req.preparedWorkerSpec.Runtime.ToolModelBindings {
		resource, err := o.modelResources.ResolveExact(
			ctx,
			resourcesvc.Actor{UserID: req.UserID},
			req.OrganizationID,
			binding.ModelBinding.ResourceID,
			resourcesvc.ResolutionRequirements{
				Modality:   binding.Modality,
				Capability: binding.Capability,
				AllowedProtocolAdapters: []string{
					binding.ModelBinding.ProtocolAdapter.String(),
				},
			},
		)
		if err != nil {
			return err
		}
		if err := validatePreparedModelBinding(binding.ModelBinding, resource); err != nil {
			return err
		}
		values, err := toolModelEnvironment(binding, resource)
		if err != nil {
			return err
		}
		if err := applyModelResourceEnv(environment, values); err != nil {
			return err
		}
	}
	if req.ModelResourceEnv == nil {
		req.ModelResourceEnv = map[string]string{}
	}
	return applyModelResourceEnv(req.ModelResourceEnv, environment)
}

func toolModelEnvironment(
	binding specdomain.ToolModelBinding,
	resource *resourcesvc.ResolvedResource,
) (map[string]string, error) {
	apiKey := modelResourceAPIKey(resource)
	baseURL := ""
	modelID := ""
	if resource != nil {
		baseURL = strings.TrimSpace(resource.Connection.BaseURL)
		modelID = strings.TrimSpace(resource.Resource.ModelID)
	}
	if apiKey == "" || baseURL == "" || modelID == "" {
		return nil, ErrMissingModelResource
	}
	return map[string]string{
		binding.Environment.APIKey:  apiKey,
		binding.Environment.BaseURL: baseURL,
		binding.Environment.ModelID: modelID,
	}, nil
}
