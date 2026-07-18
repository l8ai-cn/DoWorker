package workercreation

import (
	"context"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type modelResourceService struct {
	resolved      *resourceservice.ResolvedResource
	resolvedByID  map[int64]*resourceservice.ResolvedResource
	err           error
	calls         int
	exactCalls    int
	metadataCalls int
	actor         resourceservice.Actor
	orgID         int64
	resourceID    int64
	requirements  resourceservice.ResolutionRequirements
}

func (service *modelResourceService) ResolveExact(
	_ context.Context,
	actor resourceservice.Actor,
	orgID, resourceID int64,
	requirements resourceservice.ResolutionRequirements,
) (*resourceservice.ResolvedResource, error) {
	service.exactCalls++
	return service.resolve(actor, orgID, resourceID, requirements)
}

func (service *modelResourceService) ResolveMetadata(
	_ context.Context,
	actor resourceservice.Actor,
	orgID, resourceID int64,
	requirements resourceservice.ResolutionRequirements,
) (*resourceservice.ResolvedResource, error) {
	service.metadataCalls++
	return service.resolve(actor, orgID, resourceID, requirements)
}

func (service *modelResourceService) resolve(
	actor resourceservice.Actor,
	orgID, resourceID int64,
	requirements resourceservice.ResolutionRequirements,
) (*resourceservice.ResolvedResource, error) {
	service.calls++
	service.actor = actor
	service.orgID = orgID
	service.resourceID = resourceID
	service.requirements = requirements
	if resolved := service.resolvedByID[resourceID]; resolved != nil {
		return resolved, service.err
	}
	return service.resolved, service.err
}

func validModelResourceService() *modelResourceService {
	return &modelResourceService{
		resolved: &resourceservice.ResolvedResource{
			Provider: resourcedomain.ProviderDefinition{
				Key:             slugkit.MustNewForTest("openai"),
				ProtocolAdapter: "openai-compatible",
			},
			Connection: resourcedomain.Connection{
				ID: 201, ProviderKey: slugkit.MustNewForTest("openai"), Revision: 9,
			},
			Resource: resourcedomain.ModelResource{
				ID: 101, ProviderConnectionID: 201, ModelID: "gpt-5", Revision: 7,
			},
		},
	}
}

func requiredModelRequirement(adapters ...string) specdomain.ModelRequirement {
	requirement := specdomain.ModelRequirement{
		Required:         true,
		ProtocolAdapters: make([]slugkit.Slug, len(adapters)),
	}
	for index, adapter := range adapters {
		requirement.ProtocolAdapters[index] = slugkit.MustNewForTest(adapter)
	}
	return requirement
}
