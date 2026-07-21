package workerdependencyartifact

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
)

func materializeDocument(
	scope control.Scope,
	definition workerdefinition.Definition,
	agentfileLayer string,
	spec workerspec.Spec,
	specDigest string,
	resolved ResolvedDependencies,
) (workerdependency.Document, error) {
	worker, err := buildWorkerSnapshot(
		definition,
		agentfileLayer,
		spec,
		specDigest,
	)
	if err != nil {
		return workerdependency.Document{}, err
	}
	document := workerdependency.Document{
		Version: workerdependency.VersionV1, OrganizationID: scope.OrganizationID,
		Namespace: scope.OrganizationSlug, Worker: worker,
		Models: workerdependency.Models{
			Primary: materializePrimaryModel(resolved.PrimaryModel),
			Tools:   materializeToolModels(resolved.ToolModels),
		},
		Repository:       materializeRepository(resolved.Repository),
		Skills:           materializeSkills(resolved.Skills),
		KnowledgeBases:   materializeKnowledgeBases(resolved.KnowledgeBases),
		RuntimeBundles:   materializeRuntimeBundles(resolved.RuntimeBundles),
		SecretReferences: materializeSecretReferences(resolved.SecretReferences),
		Placement:        materializePlacement(resolved.Placement),
	}
	return document, nil
}

func materializePrimaryModel(
	resolution *ModelResolution,
) *workerdependency.Model {
	if resolution == nil {
		return nil
	}
	model := materializeModel(*resolution)
	return &model
}

func materializeModel(resolution ModelResolution) workerdependency.Model {
	return workerdependency.Model{
		Pin:                pin(resolution.ResourceResolution),
		ResourceRevision:   resolution.ResourceRevision,
		ConnectionID:       resolution.ConnectionID,
		ConnectionRevision: resolution.ConnectionRevision,
		ProviderKey:        resolution.ProviderKey,
		ProtocolAdapter:    resolution.ProtocolAdapter,
		ModelID:            resolution.ModelID, BaseURL: resolution.BaseURL,
		Modalities: append([]airesource.Modality{}, resolution.Modalities...),
		Capabilities: append(
			[]airesource.Capability{},
			resolution.Capabilities...,
		),
	}
}

func materializeToolModels(
	resolutions []ToolModelResolution,
) []workerdependency.ToolModel {
	result := make([]workerdependency.ToolModel, len(resolutions))
	for index, resolution := range resolutions {
		result[index] = workerdependency.ToolModel{
			Binding: reference(resolution.Binding), Role: resolution.Role,
			Model:    materializeModel(resolution.Model),
			Modality: resolution.Modality, Capability: resolution.Capability,
			Environment: workerdependency.ToolModelEnvironment{
				APIKeyTarget:  resolution.Environment.APIKeyTarget,
				BaseURLTarget: resolution.Environment.BaseURLTarget,
				ModelIDTarget: resolution.Environment.ModelIDTarget,
			},
		}
	}
	return result
}

func pin(resolution ResourceResolution) workerdependency.ResourcePin {
	return workerdependency.ResourcePin{
		Reference: reference(resolution.reference),
		DomainID:  resolution.domainID,
	}
}

func reference(resolved control.ResolvedReference) resource.Reference {
	return resource.Reference{
		APIVersion: resolved.APIVersion, Kind: resolved.Kind,
		Namespace: resolved.Namespace, Name: resolved.Name, UID: resolved.UID,
		Revision: resolved.Revision, Digest: resolved.Digest,
	}
}
