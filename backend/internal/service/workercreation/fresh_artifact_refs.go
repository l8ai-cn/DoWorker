package workercreation

import (
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func buildFreshArtifactReferences(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs ArtifactReferences,
	spec specdomain.Spec,
	models *modelResolver,
	runtime *runtimeCatalogResolver,
	workspace *workspaceResolver,
) (ArtifactReferences, error) {
	if err := slugkit.Validate(refsNamespace(refs).String()); err == nil {
		return refs, nil
	}
	if refs.ToolBindings == nil {
		refs.ToolBindings = map[string]control.ResolvedReference{}
	}
	if refs.ToolModels == nil {
		refs.ToolModels = map[string]control.ResolvedReference{}
	}
	if err := appendFreshModelReferences(scope, namespace, &refs, spec, models); err != nil {
		return ArtifactReferences{}, err
	}
	if err := appendFreshRuntimeReferences(scope, namespace, &refs, runtime); err != nil {
		return ArtifactReferences{}, err
	}
	if err := appendFreshWorkspaceReferences(scope, namespace, &refs, spec, workspace); err != nil {
		return ArtifactReferences{}, err
	}
	return refs, nil
}

func refsNamespace(refs ArtifactReferences) slugkit.Slug {
	if len(refs.AllPlanReferences) == 0 {
		return ""
	}
	return refs.AllPlanReferences[0].Namespace
}
