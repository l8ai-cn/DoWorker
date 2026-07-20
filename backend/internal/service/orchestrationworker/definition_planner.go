package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

const DefinitionSchemaRevision = "orchestration-definition-v1"

type DefinitionApplyArtifact struct {
	WorkerSpecSnapshotID int64 `json:"workerSpecSnapshotId"`
}

type DefinitionPlanner struct {
	meta      resource.TypeMeta
	resolver  DefinitionResolver
	goalLoops GoalLoopSlugChecker
}

func NewDefinitionPlanner(
	kind string,
	resolver DefinitionResolver,
	goalLoops GoalLoopSlugChecker,
) (*DefinitionPlanner, error) {
	if !isDefinitionKind(kind) || resolver == nil || goalLoops == nil {
		return nil, fmt.Errorf(
			"%w: invalid definition planner dependencies",
			controlservice.ErrUnavailable,
		)
	}
	return &DefinitionPlanner{
		meta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		resolver: resolver, goalLoops: goalLoops,
	}, nil
}

func (planner *DefinitionPlanner) TypeMeta() resource.TypeMeta {
	if planner == nil {
		return resource.TypeMeta{}
	}
	return planner.meta
}

func (planner *DefinitionPlanner) References(
	value any,
) ([]controlservice.DraftReference, error) {
	if planner == nil {
		return nil, control.ErrCorrupt
	}
	return definitionReferences(planner.meta.Kind, value)
}

func (planner *DefinitionPlanner) Plan(
	ctx context.Context,
	input controlservice.TargetPlanInput,
) (controlservice.TargetPlanOutput, error) {
	if planner == nil || planner.resolver == nil ||
		input.Manifest.TypeMeta != planner.meta ||
		!matchesDefinitionSpec(planner.meta.Kind, input.TypedSpec) {
		return controlservice.TargetPlanOutput{}, control.ErrCorrupt
	}
	if input.Operation == control.PlanOperationUpdate {
		switch planner.meta.Kind {
		case resource.KindWorker:
			return createOnlyIssue(
				"worker-is-create-only",
				"Worker resources are one-shot and cannot be updated.",
			), nil
		case resource.KindGoalLoop:
			return createOnlyIssue(
				"goal-loop-is-create-only",
				"GoalLoop resources are one-shot and cannot be updated.",
			), nil
		}
	}
	if planner.meta.Kind == resource.KindGoalLoop {
		exists, err := planner.goalLoops.ExistsSlug(
			ctx,
			input.Scope.OrganizationID,
			input.Manifest.Metadata.Name.String(),
		)
		if err != nil {
			return controlservice.TargetPlanOutput{}, err
		}
		if exists {
			return controlservice.TargetPlanOutput{
				Issues: []control.PlanIssue{{
					Severity: control.PlanIssueBlocking,
					Path:     "/metadata/name",
					Code:     "goal-loop-name-already-exists",
					Message:  "A GoalLoop with this name already exists.",
				}},
			}, nil
		}
	}
	if planner.meta.Kind == resource.KindPrompt {
		artifact, err := control.CanonicalJSONObject(input.Manifest.Spec)
		if err != nil {
			return controlservice.TargetPlanOutput{}, control.ErrCorrupt
		}
		return controlservice.TargetPlanOutput{
			ArtifactKind:    "PromptSpec",
			ArtifactJSON:    artifact,
			OptionsRevision: DefinitionSchemaRevision,
			Issues:          []control.PlanIssue{},
		}, nil
	}
	pins, err := newPinnedReferenceIndex(input.Scope, input.ResolvedReferences)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	workerRef := definitionWorkerTemplateReference(input.TypedSpec)
	worker, err := pins.resolve(workerRef)
	if err != nil {
		return controlservice.TargetPlanOutput{}, control.ErrCorrupt
	}
	snapshotID, err := planner.resolver.ResolveWorkerSpecSnapshotID(
		ctx,
		input.Scope,
		worker,
	)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	issues, err := planner.promptInputIssues(ctx, input, pins)
	if err != nil {
		return controlservice.TargetPlanOutput{}, err
	}
	artifact, err := definitionApplyArtifact(
		planner.meta.Kind,
		snapshotID,
		input.TypedSpec,
	)
	if err != nil {
		return controlservice.TargetPlanOutput{}, control.ErrCorrupt
	}
	return controlservice.TargetPlanOutput{
		ArtifactKind:    planner.meta.Kind + "Apply",
		ArtifactJSON:    artifact,
		OptionsRevision: DefinitionSchemaRevision,
		Issues:          issues,
	}, nil
}

func createOnlyIssue(code string, message string) controlservice.TargetPlanOutput {
	return controlservice.TargetPlanOutput{
		Issues: []control.PlanIssue{{
			Severity: control.PlanIssueBlocking,
			Path:     "/",
			Code:     code,
			Message:  message,
		}},
	}
}

func isDefinitionKind(kind string) bool {
	switch kind {
	case resource.KindPrompt, resource.KindWorker, resource.KindExpert,
		resource.KindWorkflow, resource.KindGoalLoop:
		return true
	default:
		return false
	}
}

var _ controlservice.TargetPlanner = (*DefinitionPlanner)(nil)
