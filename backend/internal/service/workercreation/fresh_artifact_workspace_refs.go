package workercreation

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func appendFreshWorkspaceReferences(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs *ArtifactReferences,
	spec specdomain.Spec,
	workspace *workspaceResolver,
) error {
	if refs.RuntimeBundles == nil {
		refs.RuntimeBundles = map[int64]control.ResolvedReference{}
	}
	if refs.ConfigBundles == nil {
		refs.ConfigBundles = map[int64]control.ResolvedReference{}
	}
	if refs.SecretBundles == nil {
		refs.SecretBundles = map[string]control.ResolvedReference{}
	}
	if refs.Skills == nil {
		refs.Skills = map[int64]control.ResolvedReference{}
	}
	if refs.KnowledgeBases == nil {
		refs.KnowledgeBases = map[int64]control.ResolvedReference{}
	}
	if err := appendFreshRepositoryReference(scope, namespace, refs, spec, workspace); err != nil {
		return err
	}
	packages := skillPackageIndex(spec.Workspace.SkillPackages)
	for _, id := range spec.Workspace.SkillIDs {
		packageBinding, err := requiredSkillPackage(packages, id)
		if err != nil {
			return err
		}
		ref, err := freshNamedDomainReference(
			scope,
			namespace,
			resource.KindSkill,
			packageBinding.Slug,
			id,
			map[string]any{
				"skill_id": id, "version": packageBinding.Version,
				"content_sha":  packageBinding.ContentSHA,
				"storage_key":  packageBinding.StorageKey,
				"package_size": packageBinding.PackageSize,
			},
		)
		if err != nil {
			return err
		}
		refs.Skills[id] = ref
		refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	}
	for _, mount := range spec.Workspace.KnowledgeMounts {
		id := mount.KnowledgeBaseID
		ref, err := freshDomainReference(
			scope,
			namespace,
			resource.KindKnowledgeBase,
			"knowledge-base",
			id,
			map[string]any{"knowledge_base_id": id, "mode": string(mount.Mode)},
		)
		if err != nil {
			return err
		}
		refs.KnowledgeBases[id] = ref
		refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	}
	for _, id := range spec.Workspace.EnvBundleIDs {
		ref, err := freshEnvBundleReference(scope, namespace, workspace.envBundles[int64(id)])
		if err != nil {
			return err
		}
		refs.RuntimeBundles[int64(id)] = ref
		refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	}
	for field, secret := range spec.TypeConfig.SecretRefs {
		ref, err := freshEnvBundleReference(scope, namespace, workspace.envBundles[secret.ID])
		if err != nil {
			return err
		}
		refs.SecretBundles[field] = ref
		refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	}
	for _, binding := range spec.Workspace.ConfigDocumentBindings {
		ref, err := freshEnvBundleReference(
			scope,
			namespace,
			workspace.envBundles[binding.ConfigBundleID],
		)
		if err != nil {
			return err
		}
		refs.ConfigBundles[binding.ConfigBundleID] = ref
		refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	}
	return nil
}

func appendFreshRepositoryReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs *ArtifactReferences,
	spec specdomain.Spec,
	workspace *workspaceResolver,
) error {
	if spec.Workspace.RepositoryID == nil {
		return nil
	}
	id := *spec.Workspace.RepositoryID
	repository := workspace.resolvedRepository(spec.Workspace.RepositoryID)
	if repository == nil {
		return fmt.Errorf("worker dependency repository %d was not resolved", id)
	}
	ref, err := freshDomainReference(
		scope,
		namespace,
		resource.KindRepository,
		"repository",
		id,
		map[string]any{
			"repository_id":   id,
			"repository_slug": repository.Slug,
			"branch":          spec.Workspace.Branch,
		},
	)
	if err != nil {
		return err
	}
	refs.Repository = &ref
	refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	return nil
}

func freshEnvBundleReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	bundle *envbundle.EnvBundle,
) (control.ResolvedReference, error) {
	if bundle == nil {
		return control.ResolvedReference{}, fmt.Errorf("worker dependency environment bundle was not resolved")
	}
	name, err := slugkit.NewFromTrusted(bundle.Name)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	revision := positiveDigestRevision(fmt.Sprintf("%d:%s", bundle.ID, bundle.UpdatedAt.UTC()))
	return freshResolvedReference(
		scope,
		namespace,
		resource.KindEnvironmentBundle,
		name,
		revision,
		map[string]any{"bundle_id": bundle.ID, "kind": bundle.Kind},
	)
}
