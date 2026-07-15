package orchestrationworker

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func resolveWorkerWorkspace(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateWorkspaceSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (workerspec.Workspace, error) {
	repositoryID, err := resolveOptionalEntityID(
		ctx,
		scope,
		spec.RepositoryRef,
		pins,
		bindings,
	)
	if err != nil {
		return workerspec.Workspace{}, err
	}
	skillIDs, err := resolveEntityIDs(
		ctx,
		scope,
		spec.SkillRefs,
		pins,
		bindings,
	)
	if err != nil {
		return workerspec.Workspace{}, err
	}
	knowledge, err := resolveKnowledgeMounts(
		ctx,
		scope,
		spec.KnowledgeMounts,
		pins,
		bindings,
	)
	if err != nil {
		return workerspec.Workspace{}, err
	}
	environmentIDs, err := resolveEntityIDs(
		ctx,
		scope,
		spec.EnvironmentBundleRefs,
		pins,
		bindings,
	)
	if err != nil {
		return workerspec.Workspace{}, err
	}
	configIDs, err := resolveEntityIDs(
		ctx,
		scope,
		spec.ConfigBundleRefs,
		pins,
		bindings,
	)
	if err != nil {
		return workerspec.Workspace{}, err
	}
	var repository *int64
	if spec.RepositoryRef != nil {
		repository = &repositoryID
	}
	envBundles := make([]workerspec.RuntimeEnvBundleID, len(environmentIDs))
	for index, id := range environmentIDs {
		envBundles[index] = workerspec.RuntimeEnvBundleID(id)
	}
	return workerspec.Workspace{
		RepositoryID: repository, Branch: spec.Branch, SkillIDs: skillIDs,
		KnowledgeMounts: knowledge, EnvBundleIDs: envBundles,
		ConfigBundleIDs: configIDs, Instructions: spec.Instructions,
		InitialTask: "",
	}, nil
}

func resolveKnowledgeMounts(
	ctx context.Context,
	scope control.Scope,
	mounts []resource.WorkerTemplateKnowledgeMount,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) ([]workerspec.KnowledgeMount, error) {
	resolved := make([]workerspec.KnowledgeMount, len(mounts))
	for index, mount := range mounts {
		id, err := resolveEntityID(ctx, scope, mount.Ref, pins, bindings)
		if err != nil {
			return nil, err
		}
		resolved[index] = workerspec.KnowledgeMount{
			KnowledgeBaseID: id,
			Mode:            mount.Mode,
		}
	}
	return resolved, nil
}

func resolveEntityIDs(
	ctx context.Context,
	scope control.Scope,
	references []resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) ([]int64, error) {
	ids := make([]int64, len(references))
	for index, reference := range references {
		id, err := resolveEntityID(ctx, scope, reference, pins, bindings)
		if err != nil {
			return nil, err
		}
		ids[index] = id
	}
	return ids, nil
}
