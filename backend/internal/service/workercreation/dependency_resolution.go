package workercreation

import (
	"context"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
)

func buildResolvedDependencies(
	ctx context.Context,
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
		ctx,
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
