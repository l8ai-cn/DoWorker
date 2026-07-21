package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

type ResourceBindingResolver struct {
	registry   *resource.Registry
	repository controlservice.Repository
	authorizer controlservice.Authorizer
}

func NewResourceBindingResolver(
	registry *resource.Registry,
	repository controlservice.Repository,
	authorizer controlservice.Authorizer,
) (*ResourceBindingResolver, error) {
	if registry == nil || repository == nil || authorizer == nil {
		return nil, fmt.Errorf(
			"%w: incomplete resource binding resolver",
			controlservice.ErrUnavailable,
		)
	}
	for _, kind := range resourceBindingKinds() {
		if !registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		}) {
			return nil, fmt.Errorf(
				"%w: resource binding schema %s is not registered",
				controlservice.ErrUnavailable,
				kind,
			)
		}
	}
	return &ResourceBindingResolver{
		registry: registry, repository: repository, authorizer: authorizer,
	}, nil
}

func (resolver *ResourceBindingResolver) ResolveEntityID(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	if !isEntityBindingKind(reference.Kind) {
		return 0, control.ErrCorrupt
	}
	binding, err := resolver.loadBindingRevision(ctx, scope, reference)
	if err != nil {
		return 0, err
	}
	if len(binding.revision.ResolvedReferences) != 0 {
		return 0, control.ErrCorrupt
	}
	id, ok := entityBindingID(binding.spec)
	if !ok || id <= 0 {
		return 0, control.ErrCorrupt
	}
	return id, nil
}

func (resolver *ResourceBindingResolver) ResolveToolModel(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (ToolModelBindingResolution, error) {
	if reference.Kind != resource.KindToolBinding {
		return ToolModelBindingResolution{}, control.ErrCorrupt
	}
	binding, err := resolver.loadBindingRevision(ctx, scope, reference)
	if err != nil {
		return ToolModelBindingResolution{}, err
	}
	spec, ok := binding.spec.(*resource.ToolBindingSpec)
	if !ok || spec == nil || len(binding.revision.ResolvedReferences) != 1 {
		return ToolModelBindingResolution{}, control.ErrCorrupt
	}
	pins, err := newPinnedReferenceIndex(
		scope,
		binding.revision.ResolvedReferences,
	)
	if err != nil {
		return ToolModelBindingResolution{}, control.ErrCorrupt
	}
	model, err := pins.resolve(spec.ModelRef)
	if err != nil || model.Kind != resource.KindModelBinding {
		return ToolModelBindingResolution{}, control.ErrCorrupt
	}
	id, err := resolver.ResolveEntityID(ctx, scope, model)
	if err != nil {
		return ToolModelBindingResolution{}, err
	}
	return ToolModelBindingResolution{
		Binding: reference, ModelBinding: model, ModelResourceID: id,
	}, nil
}

func (resolver *ResourceBindingResolver) ResolveToolModelResourceID(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	resolved, err := resolver.ResolveToolModel(ctx, scope, reference)
	if err != nil {
		return 0, err
	}
	return resolved.ModelResourceID, nil
}

func entityBindingID(spec any) (int64, bool) {
	switch value := spec.(type) {
	case *resource.ModelBindingSpec:
		return value.ResourceID, true
	case *resource.RepositoryBindingSpec:
		return value.RepositoryID, true
	case *resource.SkillBindingSpec:
		return value.SkillID, true
	case *resource.KnowledgeBaseBindingSpec:
		return value.KnowledgeBaseID, true
	case *resource.EnvironmentBundleBindingSpec:
		return value.EnvironmentBundleID, true
	case *resource.ComputeTargetBindingSpec:
		return value.ComputeTargetID, true
	case *resource.ResourceProfileBindingSpec:
		return value.ResourceProfileID, true
	default:
		return 0, false
	}
}

func (resolver *ResourceBindingResolver) ResolveWorkerSpecSnapshotID(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	if reference.Kind != resource.KindWorkerTemplate {
		return 0, control.ErrCorrupt
	}
	loaded, err := resolver.loadBindingRevision(ctx, scope, reference)
	if err != nil {
		return 0, err
	}
	if loaded.revision.WorkerSpecSnapshotID <= 0 {
		return 0, control.ErrCorrupt
	}
	return loaded.revision.WorkerSpecSnapshotID, nil
}

func (resolver *ResourceBindingResolver) ResolvePromptSpec(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (resource.PromptSpec, error) {
	if reference.Kind != resource.KindPrompt {
		return resource.PromptSpec{}, control.ErrCorrupt
	}
	loaded, err := resolver.loadBindingRevision(ctx, scope, reference)
	if err != nil {
		return resource.PromptSpec{}, err
	}
	spec, ok := loaded.spec.(*resource.PromptSpec)
	if !ok || spec == nil {
		return resource.PromptSpec{}, control.ErrCorrupt
	}
	return *spec, nil
}

func resourceBindingKinds() []string {
	return append(entityBindingKinds(),
		resource.KindToolBinding,
	)
}

func entityBindingKinds() []string {
	return []string{
		resource.KindModelBinding,
		resource.KindRepository,
		resource.KindSkill,
		resource.KindKnowledgeBase,
		resource.KindEnvironmentBundle,
		resource.KindComputeTarget,
		resource.KindResourceProfile,
	}
}

func isEntityBindingKind(kind string) bool {
	for _, candidate := range entityBindingKinds() {
		if kind == candidate {
			return true
		}
	}
	return false
}

var _ BindingResolver = (*ResourceBindingResolver)(nil)
var _ DefinitionResolver = (*ResourceBindingResolver)(nil)
