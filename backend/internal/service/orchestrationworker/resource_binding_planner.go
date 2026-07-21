package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

const BindingSchemaRevision = "orchestration-binding-v1"

type ResourceBindingPlanner struct {
	meta resource.TypeMeta
}

func NewResourceBindingPlanner(
	kind string,
) (*ResourceBindingPlanner, error) {
	if !isBindingResourceKind(kind) {
		return nil, fmt.Errorf(
			"%w: unsupported binding resource kind",
			controlservice.ErrUnavailable,
		)
	}
	return &ResourceBindingPlanner{meta: resource.TypeMeta{
		APIVersion: resource.APIVersionV1Alpha1,
		Kind:       kind,
	}}, nil
}

func (planner *ResourceBindingPlanner) TypeMeta() resource.TypeMeta {
	if planner == nil {
		return resource.TypeMeta{}
	}
	return planner.meta
}

func (planner *ResourceBindingPlanner) References(
	value any,
) ([]controlservice.DraftReference, error) {
	if planner == nil || !matchesBindingSpec(planner.meta.Kind, value) {
		return nil, control.ErrCorrupt
	}
	if planner.meta.Kind != resource.KindToolBinding {
		return []controlservice.DraftReference{}, nil
	}
	spec := value.(*resource.ToolBindingSpec)
	return []controlservice.DraftReference{{
		Path: "/spec/modelRef", Reference: spec.ModelRef,
	}}, nil
}

func (planner *ResourceBindingPlanner) Plan(
	_ context.Context,
	input controlservice.TargetPlanInput,
) (controlservice.TargetPlanOutput, error) {
	if planner == nil ||
		input.Manifest.TypeMeta != planner.meta ||
		!matchesBindingSpec(planner.meta.Kind, input.TypedSpec) {
		return controlservice.TargetPlanOutput{}, control.ErrCorrupt
	}
	artifact, err := control.CanonicalJSONObject(input.Manifest.Spec)
	if err != nil {
		return controlservice.TargetPlanOutput{}, control.ErrCorrupt
	}
	return controlservice.TargetPlanOutput{
		ArtifactKind:    planner.meta.Kind + "Spec",
		ArtifactJSON:    artifact,
		OptionsRevision: BindingSchemaRevision,
		Issues:          []control.PlanIssue{},
	}, nil
}

func isBindingResourceKind(kind string) bool {
	for _, candidate := range resourceBindingKinds() {
		if kind == candidate {
			return true
		}
	}
	return false
}

func matchesBindingSpec(kind string, value any) bool {
	switch kind {
	case resource.KindModelBinding:
		_, ok := value.(*resource.ModelBindingSpec)
		return ok
	case resource.KindRepository:
		_, ok := value.(*resource.RepositoryBindingSpec)
		return ok
	case resource.KindSkill:
		_, ok := value.(*resource.SkillBindingSpec)
		return ok
	case resource.KindKnowledgeBase:
		_, ok := value.(*resource.KnowledgeBaseBindingSpec)
		return ok
	case resource.KindEnvironmentBundle:
		_, ok := value.(*resource.EnvironmentBundleBindingSpec)
		return ok
	case resource.KindComputeTarget:
		_, ok := value.(*resource.ComputeTargetBindingSpec)
		return ok
	case resource.KindResourceProfile:
		_, ok := value.(*resource.ResourceProfileBindingSpec)
		return ok
	case resource.KindToolBinding:
		_, ok := value.(*resource.ToolBindingSpec)
		return ok
	default:
		return false
	}
}

var _ controlservice.TargetPlanner = (*ResourceBindingPlanner)(nil)
