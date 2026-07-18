package workercreation

import (
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
)

func buildResolvedDependencies(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	definition workerdefinition.Definition,
	models *modelResolver,
	runtime *runtimeCatalogResolver,
	workspace *workspaceResolver,
) (workerdependencyartifact.ResolvedDependencies, error) {
	primary, err := buildPrimaryModelResolution(scope, refs, spec, models)
	if err != nil {
		return workerdependencyartifact.ResolvedDependencies{}, err
	}
	tools, err := buildToolModelResolutions(scope, refs, spec, models)
	if err != nil {
		return workerdependencyartifact.ResolvedDependencies{}, err
	}
	workspaceDeps, err := buildWorkspaceDependencies(
		scope,
		refs,
		spec,
		definition,
		workspace,
	)
	if err != nil {
		return workerdependencyartifact.ResolvedDependencies{}, err
	}
	placement, err := buildPlacementResolution(scope, refs, spec, runtime)
	if err != nil {
		return workerdependencyartifact.ResolvedDependencies{}, err
	}
	return workerdependencyartifact.ResolvedDependencies{
		PrimaryModel:     primary,
		ToolModels:       tools,
		Repository:       workspaceDeps.repository,
		Skills:           workspaceDeps.skills,
		KnowledgeBases:   workspaceDeps.knowledge,
		RuntimeBundles:   workspaceDeps.bundles,
		SecretReferences: workspaceDeps.secrets,
		Placement:        placement,
	}, nil
}
