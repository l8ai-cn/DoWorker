package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
)

type WorkerTemplatePlanner struct {
	bindings BindingResolver
	compiler WorkerCompiler
}

func NewWorkerTemplatePlanner(
	bindings BindingResolver,
	compiler WorkerCompiler,
) (*WorkerTemplatePlanner, error) {
	if bindings == nil || compiler == nil || compiler.Revision() == "" {
		return nil, fmt.Errorf(
			"%w: incomplete WorkerTemplate planner",
			controlservice.ErrUnavailable,
		)
	}
	return &WorkerTemplatePlanner{bindings: bindings, compiler: compiler}, nil
}

func (*WorkerTemplatePlanner) TypeMeta() resource.TypeMeta {
	return resource.TypeMeta{
		APIVersion: resource.APIVersionV1Alpha1,
		Kind:       resource.KindWorkerTemplate,
	}
}

func (planner *WorkerTemplatePlanner) References(
	value any,
) ([]controlservice.DraftReference, error) {
	spec, ok := value.(*resource.WorkerTemplateSpec)
	if !ok || spec == nil {
		return nil, fmt.Errorf("%w: invalid WorkerTemplate spec", control.ErrCorrupt)
	}
	return workerTemplateReferences(*spec), nil
}

func (planner *WorkerTemplatePlanner) Plan(
	ctx context.Context,
	input controlservice.TargetPlanInput,
) (controlservice.TargetPlanOutput, error) {
	if planner == nil || planner.bindings == nil || planner.compiler == nil {
		return controlservice.TargetPlanOutput{}, controlservice.ErrUnavailable
	}
	spec, ok := input.TypedSpec.(*resource.WorkerTemplateSpec)
	if !ok || spec == nil {
		return controlservice.TargetPlanOutput{}, fmt.Errorf(
			"%w: invalid WorkerTemplate spec",
			control.ErrCorrupt,
		)
	}
	if spec.OptionsRevision != planner.compiler.Revision() {
		return controlservice.TargetPlanOutput{}, controlservice.ErrStaleOptions
	}
	pins, err := newPinnedReferenceIndex(input.Scope, input.ResolvedReferences)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	draft, err := buildWorkerTemplateDraft(
		ctx,
		input.Scope,
		*spec,
		pins,
		planner.bindings,
	)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	compilation, err := planner.compiler.Compile(ctx, input.Scope, draft)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	return controlservice.TargetPlanOutput{
		ArtifactKind:    workerdependencyartifact.PlanArtifactKind,
		ArtifactJSON:    compilation.ArtifactJSON,
		OptionsRevision: planner.compiler.Revision(),
		Issues:          compilation.Issues,
	}, nil
}
