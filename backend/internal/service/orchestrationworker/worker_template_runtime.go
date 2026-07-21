package orchestrationworker

import (
	"context"
	"sort"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

func resolveWorkerRuntime(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateRuntimeSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (specservice.RuntimeSelection, error) {
	computeTargetID, err := resolveEntityID(
		ctx,
		scope,
		spec.ComputeTargetRef,
		pins,
		bindings,
	)
	if err != nil {
		return specservice.RuntimeSelection{}, err
	}
	resourceProfileID, err := resolveOptionalEntityID(
		ctx,
		scope,
		spec.ResourceProfileRef,
		pins,
		bindings,
	)
	if err != nil {
		return specservice.RuntimeSelection{}, err
	}
	return specservice.RuntimeSelection{
		RuntimeImageID: spec.RuntimeImageID, PlacementPolicy: spec.PlacementPolicy,
		ComputeTargetID: computeTargetID, DeploymentMode: spec.DeploymentMode,
		ResourceProfileID: resourceProfileID,
		CustomResources:   cloneResourceLimits(spec.CustomResources),
	}, nil
}

func resolveWorkerTypeConfig(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateTypeConfigSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (specdomain.TypeConfig, error) {
	fields := make([]string, 0, len(spec.SecretRefs))
	for field := range spec.SecretRefs {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	secretRefs := make(map[string]specdomain.SecretReference, len(fields))
	for _, field := range fields {
		id, err := resolveEntityID(
			ctx,
			scope,
			spec.SecretRefs[field],
			pins,
			bindings,
		)
		if err != nil {
			return specdomain.TypeConfig{}, err
		}
		secretRefs[field] = specdomain.SecretReference{
			Kind: "env-bundle",
			ID:   id,
		}
	}
	return specdomain.TypeConfig{
		SchemaVersion: spec.SchemaVersion, Values: cloneWorkerValues(spec.Values),
		SecretRefs: secretRefs, InteractionMode: spec.InteractionMode,
		AutomationLevel: spec.AutomationLevel,
	}, nil
}

func resolveToolModels(
	ctx context.Context,
	scope control.Scope,
	references map[string]resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (map[string]int64, error) {
	roles := make([]string, 0, len(references))
	for role := range references {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	models := make(map[string]int64, len(roles))
	for _, role := range roles {
		pinned, err := pins.resolve(references[role])
		if err != nil {
			return nil, err
		}
		resolved, err := bindings.ResolveToolModel(ctx, scope, pinned)
		if err != nil {
			return nil, err
		}
		models[role] = resolved.ModelResourceID
	}
	return models, nil
}

func cloneResourceLimits(
	resources *specdomain.ResourceRequestsLimits,
) *specdomain.ResourceRequestsLimits {
	if resources == nil {
		return nil
	}
	cloned := *resources
	if resources.GPURequest != nil {
		value := *resources.GPURequest
		cloned.GPURequest = &value
	}
	if resources.GPULimit != nil {
		value := *resources.GPULimit
		cloned.GPULimit = &value
	}
	return &cloned
}
