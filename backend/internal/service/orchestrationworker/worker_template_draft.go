package orchestrationworker

import (
	"context"
	"maps"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func buildWorkerTemplateDraft(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (workercreation.Draft, error) {
	modelID, err := resolveOptionalEntityID(
		ctx,
		scope,
		spec.ModelRef,
		pins,
		bindings,
	)
	if err != nil {
		return workercreation.Draft{}, err
	}
	toolModels, err := resolveToolModels(
		ctx,
		scope,
		spec.ToolRefs,
		pins,
		bindings,
	)
	if err != nil {
		return workercreation.Draft{}, err
	}
	runtime, err := resolveWorkerRuntime(
		ctx,
		scope,
		spec.Runtime,
		pins,
		bindings,
	)
	if err != nil {
		return workercreation.Draft{}, err
	}
	typeConfig, err := resolveWorkerTypeConfig(
		ctx,
		scope,
		spec.TypeConfig,
		pins,
		bindings,
	)
	if err != nil {
		return workercreation.Draft{}, err
	}
	workspace, err := resolveWorkerWorkspace(
		ctx,
		scope,
		spec.Workspace,
		pins,
		bindings,
	)
	if err != nil {
		return workercreation.Draft{}, err
	}
	return workercreation.Draft{
		OptionsRevision: spec.OptionsRevision,
		WorkerSpec: specservice.Draft{
			ModelResourceID: modelID, ToolModelResourceIDs: toolModels,
			WorkerTypeSlug: spec.WorkerType, Runtime: runtime,
			TypeConfig: typeConfig, Workspace: workspace,
			Lifecycle: specdomain.Lifecycle{
				TerminationPolicy:  spec.Lifecycle.TerminationPolicy,
				IdleTimeoutMinutes: spec.Lifecycle.IdleTimeoutMinutes,
			},
			Metadata: specdomain.Metadata{Alias: spec.Metadata.Alias},
		},
	}, nil
}

func cloneWorkerValues(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	return maps.Clone(values)
}
