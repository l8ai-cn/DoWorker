package workercreation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func buildPrimaryModelResolution(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	models *modelResolver,
) (*workerdependencyartifact.ModelResolution, error) {
	binding := spec.Runtime.ModelBinding
	if binding.ResourceID == 0 {
		return nil, nil
	}
	if refs.PrimaryModel == nil {
		return nil, fmt.Errorf("WorkerTemplate artifact primary model reference is missing")
	}
	model, err := buildModelResolution(scope, *refs.PrimaryModel, binding.ResourceID, models)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func buildToolModelResolutions(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	models *modelResolver,
) ([]workerdependencyartifact.ToolModelResolution, error) {
	bindings := append([]workerspec.ToolModelBinding{}, spec.Runtime.ToolModelBindings...)
	sort.Slice(bindings, func(left, right int) bool {
		return bindings[left].Role < bindings[right].Role
	})
	result := make([]workerdependencyartifact.ToolModelResolution, 0, len(bindings))
	for _, binding := range bindings {
		role := binding.Role.String()
		parent, child := refs.ToolBindings[role], refs.ToolModels[role]
		if parent.Kind == "" || child.Kind == "" {
			return nil, fmt.Errorf("WorkerTemplate artifact tool model %q reference is missing", role)
		}
		model, err := buildModelResolution(
			scope,
			child,
			binding.ModelBinding.ResourceID,
			models,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, workerdependencyartifact.ToolModelResolution{
			Binding: parent, Role: binding.Role, Model: model,
			Modality: binding.Modality, Capability: binding.Capability,
			Environment: workerdependencyartifact.ToolModelEnvironmentResolution{
				APIKeyTarget:  binding.Environment.APIKey,
				BaseURLTarget: binding.Environment.BaseURL,
				ModelIDTarget: binding.Environment.ModelID,
			},
		})
	}
	return result, nil
}

func buildModelResolution(
	scope control.Scope,
	reference control.ResolvedReference,
	resourceID int64,
	models *modelResolver,
) (workerdependencyartifact.ModelResolution, error) {
	resolved, ok := models.resolvedModel(resourceID)
	if !ok {
		return workerdependencyartifact.ModelResolution{}, fmt.Errorf(
			"WorkerTemplate artifact model %d was not resolved",
			resourceID,
		)
	}
	pin, err := workerdependencyartifact.BindResourceProjection(scope, reference, resourceID)
	if err != nil {
		return workerdependencyartifact.ModelResolution{}, err
	}
	protocol, err := slugkit.NewFromTrusted(resolved.Provider.ProtocolAdapter)
	if err != nil {
		return workerdependencyartifact.ModelResolution{}, err
	}
	return modelResolutionFromResource(pin, protocol, resolved), nil
}

func modelResolutionFromResource(
	pin workerdependencyartifact.ResourceResolution,
	protocol slugkit.Slug,
	resolved *resourceservice.ResolvedResource,
) workerdependencyartifact.ModelResolution {
	return workerdependencyartifact.ModelResolution{
		ResourceResolution: pin, ResourceRevision: resolved.Resource.Revision,
		ConnectionID: resolved.Connection.ID, ConnectionRevision: resolved.Connection.Revision,
		ProviderKey: resolved.Connection.ProviderKey, ProtocolAdapter: protocol,
		ModelID:      strings.TrimSpace(resolved.Resource.ModelID),
		BaseURL:      resolved.Connection.BaseURL,
		Modalities:   append([]airesource.Modality(nil), resolved.Resource.Modalities...),
		Capabilities: append([]airesource.Capability(nil), resolved.Resource.Capabilities...),
	}
}
