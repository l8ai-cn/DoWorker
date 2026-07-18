package workercreation

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func (service *Service) buildArtifact(
	scope specservice.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	agentfileLayer string,
	models *modelResolver,
	runtime *runtimeCatalogResolver,
	workspace *workspaceResolver,
) (*workerdependencyartifact.Artifact, error) {
	if len(refs.AllPlanReferences) == 0 {
		return nil, nil
	}
	controlScope, err := artifactScope(scope, refs.AllPlanReferences)
	if err != nil {
		return nil, err
	}
	definition, err := workspace.workerDefinition(spec.Runtime.WorkerType.Slug)
	if err != nil {
		return nil, err
	}
	dependencies, err := buildResolvedDependencies(
		controlScope,
		refs,
		spec,
		definition,
		models,
		runtime,
		workspace,
	)
	if err != nil {
		return nil, err
	}
	artifact, err := workerdependencyartifact.Build(workerdependencyartifact.Input{
		Scope: controlScope, Definition: definition,
		AgentfileLayer: agentfileLayer, PlanReferences: refs.AllPlanReferences,
		WorkerSpec: spec, Dependencies: dependencies,
	})
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

func artifactScope(
	scope specservice.Scope,
	references []control.ResolvedReference,
) (control.Scope, error) {
	if scope.OrgID <= 0 || scope.UserID <= 0 || len(references) == 0 {
		return control.Scope{}, specservice.ErrResolverUnavailable
	}
	namespace := references[0].Namespace
	result := control.Scope{
		OrganizationID: scope.OrgID, OrganizationSlug: namespace,
		ActorID: scope.UserID,
	}
	for _, reference := range references {
		if reference.Namespace != namespace {
			return control.Scope{}, fmt.Errorf("WorkerTemplate artifact has mixed namespaces")
		}
		if err := reference.Validate(result); err != nil {
			return control.Scope{}, err
		}
	}
	return result, nil
}
