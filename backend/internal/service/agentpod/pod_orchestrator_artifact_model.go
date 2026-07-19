package agentpod

import (
	"context"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

type pinnedModelCredentialResolver interface {
	ResolvePinnedCredentials(
		context.Context,
		resourcesvc.Actor,
		int64,
		int64,
		int64,
	) (map[string]string, error)
}

func (o *PodOrchestrator) artifactModelResource(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	model workerdependency.Model,
) (*resourcesvc.ResolvedResource, error) {
	resolver, ok := o.modelResources.(pinnedModelCredentialResolver)
	if !ok {
		return nil, ErrModelResourceResolverUnavailable
	}
	provider, exists := resourcedomain.Provider(model.ProviderKey.String())
	if !exists {
		return nil, ErrModelResourceProviderUnsupported
	}
	credentials, err := resolver.ResolvePinnedCredentials(
		ctx,
		resourcesvc.Actor{UserID: req.UserID},
		req.OrganizationID,
		model.Pin.DomainID,
		model.ConnectionID,
	)
	if err != nil {
		return nil, err
	}
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourcedomain.Connection{
			ID:          model.ConnectionID,
			ProviderKey: model.ProviderKey,
			BaseURL:     model.BaseURL,
			Revision:    model.ConnectionRevision,
			IsEnabled:   true,
		},
		Resource: resourcedomain.ModelResource{
			ID:                   model.Pin.DomainID,
			ModelID:              model.ModelID,
			Modalities:           append([]resourcedomain.Modality{}, model.Modalities...),
			Capabilities:         append([]resourcedomain.Capability{}, model.Capabilities...),
			Revision:             model.ResourceRevision,
			IsEnabled:            true,
			ProviderConnectionID: model.ConnectionID,
		},
		Credentials: credentials,
	}, nil
}
