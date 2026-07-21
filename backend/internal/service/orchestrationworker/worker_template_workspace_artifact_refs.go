package orchestrationworker

import (
	"context"
	"sort"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

func bindWorkspaceReferences(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateWorkspaceSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	refs *workercreation.ArtifactReferences,
) error {
	if err := bindOptionalReference(
		ctx, scope, spec.RepositoryRef, pins, bindings, &refs.Repository,
	); err != nil {
		return err
	}
	if err := bindReferenceIDs(ctx, scope, spec.SkillRefs, pins, bindings, refs.Skills); err != nil {
		return err
	}
	if err := bindKnowledgeReferenceIDs(ctx, scope, spec.KnowledgeMounts, pins, bindings, refs.KnowledgeBases); err != nil {
		return err
	}
	if err := bindReferenceIDs(ctx, scope, spec.EnvironmentBundleRefs, pins, bindings, refs.RuntimeBundles); err != nil {
		return err
	}
	for _, binding := range spec.ConfigDocumentBindings {
		pinned, id, err := resolvePinnedID(
			ctx, scope, binding.ConfigBundleRef, pins, bindings,
		)
		if err != nil {
			return err
		}
		refs.ConfigBundles[id] = pinned
	}
	return nil
}

func bindSecretReferences(
	ctx context.Context,
	scope control.Scope,
	spec resource.WorkerTemplateTypeConfigSpec,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	refs *workercreation.ArtifactReferences,
) error {
	fields := make([]string, 0, len(spec.SecretRefs))
	for field := range spec.SecretRefs {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		pinned, _, err := resolvePinnedID(
			ctx, scope, spec.SecretRefs[field], pins, bindings,
		)
		if err != nil {
			return err
		}
		refs.SecretBundles[field] = pinned
	}
	return nil
}

func bindReferenceIDs(
	ctx context.Context,
	scope control.Scope,
	references []resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	target map[int64]control.ResolvedReference,
) error {
	for _, reference := range references {
		pinned, id, err := resolvePinnedID(ctx, scope, reference, pins, bindings)
		if err != nil {
			return err
		}
		target[id] = pinned
	}
	return nil
}

func bindKnowledgeReferenceIDs(
	ctx context.Context,
	scope control.Scope,
	mounts []resource.WorkerTemplateKnowledgeMount,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
	target map[int64]control.ResolvedReference,
) error {
	for _, mount := range mounts {
		pinned, id, err := resolvePinnedID(ctx, scope, mount.Ref, pins, bindings)
		if err != nil {
			return err
		}
		target[id] = pinned
	}
	return nil
}
