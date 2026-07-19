package workercreation

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func appendFreshModelReferences(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs *ArtifactReferences,
	spec specdomain.Spec,
	models *modelResolver,
) error {
	if spec.Runtime.ModelBinding.IsEmpty() {
		return nil
	}
	resolved, ok := models.resolvedModel(spec.Runtime.ModelBinding.ResourceID)
	if !ok {
		return fmt.Errorf("worker dependency primary model was not resolved")
	}
	ref, err := freshModelReference(scope, namespace, resolved)
	if err != nil {
		return err
	}
	refs.PrimaryModel = &ref
	refs.AllPlanReferences = append(refs.AllPlanReferences, ref)
	for _, binding := range spec.Runtime.ToolModelBindings {
		if err := appendFreshToolModelReference(
			scope,
			namespace,
			refs,
			binding,
			models,
		); err != nil {
			return err
		}
	}
	return nil
}

func appendFreshRuntimeReferences(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs *ArtifactReferences,
	runtime *runtimeCatalogResolver,
) error {
	_, target, profile, ok := runtime.resolvedRuntime()
	if !ok {
		return fmt.Errorf("worker dependency runtime was not resolved")
	}
	targetRef, err := freshCatalogReference(
		scope,
		namespace,
		resource.KindComputeTarget,
		target.Slug,
		runtime.catalog.Revision(),
	)
	if err != nil {
		return err
	}
	refs.ComputeTarget = &targetRef
	refs.AllPlanReferences = append(refs.AllPlanReferences, targetRef)
	if profile == nil {
		return nil
	}
	profileRef, err := freshCatalogReference(
		scope,
		namespace,
		resource.KindResourceProfile,
		profile.Slug,
		runtime.catalog.Revision(),
	)
	if err != nil {
		return err
	}
	refs.ResourceProfile = &profileRef
	refs.AllPlanReferences = append(refs.AllPlanReferences, profileRef)
	return nil
}

func freshModelReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	resolved *resourceservice.ResolvedResource,
) (control.ResolvedReference, error) {
	return freshResolvedReference(
		scope,
		namespace,
		resource.KindModelBinding,
		resolved.Resource.Identifier,
		resolved.Resource.Revision,
		map[string]any{"model_id": resolved.Resource.ModelID},
	)
}

func freshCatalogReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	kind, name, revision string,
) (control.ResolvedReference, error) {
	slug, err := slugkit.NewFromTrusted(name)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	return freshResolvedReference(
		scope,
		namespace,
		kind,
		slug,
		positiveDigestRevision(revision),
		map[string]any{"catalog_revision": revision},
	)
}

func appendFreshToolModelReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	refs *ArtifactReferences,
	binding specdomain.ToolModelBinding,
	models *modelResolver,
) error {
	resolved, ok := models.resolvedModel(binding.ModelBinding.ResourceID)
	if !ok {
		return fmt.Errorf("worker dependency tool model %q was not resolved", binding.Role)
	}
	modelRef, err := freshModelReference(scope, namespace, resolved)
	if err != nil {
		return err
	}
	toolRef, err := freshResolvedReference(
		scope,
		namespace,
		resource.KindToolBinding,
		binding.Role,
		binding.ModelBinding.ResourceRevision,
		map[string]any{
			"role":              binding.Role.String(),
			"model_resource_id": binding.ModelBinding.ResourceID,
		},
	)
	if err != nil {
		return err
	}
	role := binding.Role.String()
	refs.ToolBindings[role] = toolRef
	refs.ToolModels[role] = modelRef
	refs.AllPlanReferences = append(refs.AllPlanReferences, toolRef, modelRef)
	return nil
}
