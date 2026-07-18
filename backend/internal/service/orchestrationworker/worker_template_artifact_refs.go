package orchestrationworker

import (
	"context"
	"sort"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
)

func buildWorkerTemplateArtifactReferences(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (workercreation.ArtifactReferences, error) {
	refs := workercreation.ArtifactReferences{
		ToolBindings:      map[string]control.ResolvedReference{},
		ToolModels:        map[string]control.ResolvedReference{},
		Skills:            map[int64]control.ResolvedReference{},
		KnowledgeBases:    map[int64]control.ResolvedReference{},
		RuntimeBundles:    map[int64]control.ResolvedReference{},
		SecretBundles:     map[string]control.ResolvedReference{},
		ConfigBundles:     map[int64]control.ResolvedReference{},
		AllPlanReferences: append([]control.ResolvedReference{}, pins.all()...),
	}
	if err := bindOptionalReference(
		ctx, scope, spec.ModelRef, pins, bindings, &refs.PrimaryModel,
	); err != nil {
		return workercreation.ArtifactReferences{}, err
	}
	if err := bindToolReferences(ctx, scope, spec.ToolRefs, pins, bindings, &refs); err != nil {
		return workercreation.ArtifactReferences{}, err
	}
	if err := bindRuntimeReferences(ctx, scope, spec.Runtime, pins, bindings, &refs); err != nil {
		return workercreation.ArtifactReferences{}, err
	}
	if err := bindWorkspaceReferences(ctx, scope, spec.Workspace, pins, bindings, &refs); err != nil {
		return workercreation.ArtifactReferences{}, err
	}
	if err := bindSecretReferences(ctx, scope, spec.TypeConfig, pins, bindings, &refs); err != nil {
		return workercreation.ArtifactReferences{}, err
	}
	return refs, nil
}

func bindOptionalReference(
	ctx context.Context,
	scope control.Scope,
	reference *resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	target **control.ResolvedReference,
) error {
	if reference == nil {
		return nil
	}
	pinned, err := pins.resolve(*reference)
	if err != nil {
		return err
	}
	id, err := bindings.ResolveEntityID(ctx, scope, pinned)
	if err != nil {
		return err
	}
	if id <= 0 {
		return control.ErrCorrupt
	}
	resolved := pinned
	*target = &resolved
	return nil
}

func bindRequiredReference(
	ctx context.Context,
	scope control.Scope,
	reference resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	target **control.ResolvedReference,
) error {
	return bindOptionalReference(ctx, scope, &reference, pins, bindings, target)
}

func bindToolReferences(
	ctx context.Context,
	scope control.Scope,
	references map[string]resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	refs *workercreation.ArtifactReferences,
) error {
	roles := make([]string, 0, len(references))
	for role := range references {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		pinned, err := pins.resolve(references[role])
		if err != nil {
			return err
		}
		resolved, err := bindings.ResolveToolModel(ctx, scope, pinned)
		if err != nil {
			return err
		}
		refs.ToolBindings[role] = resolved.Binding
		refs.ToolModels[role] = resolved.ModelBinding
	}
	return nil
}

func bindRuntimeReferences(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateRuntimeSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	refs *workercreation.ArtifactReferences,
) error {
	if err := bindRequiredReference(
		ctx, scope, spec.ComputeTargetRef, pins, bindings, &refs.ComputeTarget,
	); err != nil {
		return err
	}
	return bindOptionalReference(
		ctx, scope, spec.ResourceProfileRef, pins, bindings, &refs.ResourceProfile,
	)
}
