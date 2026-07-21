package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"errors"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type validatedDraft struct {
	scope     control.Scope
	result    ValidationResult
	manifest  orchestrationresource.Manifest
	typedSpec any
	planner   TargetPlanner
	head      *control.ResourceHead
}

func (service *Service) Validate(
	ctx context.Context,
	request ValidateRequest,
) (ValidationResult, error) {
	draft, err := service.validateDraft(ctx, request)
	return draft.result, err
}

func (service *Service) validateDraft(
	ctx context.Context,
	request ValidateRequest,
) (validatedDraft, error) {
	if service == nil {
		return validatedDraft{}, ErrUnavailable
	}
	if err := request.Scope.Validate(); err != nil {
		return validatedDraft{}, err
	}
	manifest, err := decodeResourceSource(request.Source)
	if err != nil {
		return invalidDraftResult(), nil
	}
	target := control.ResourceTarget{
		TypeMeta:  manifest.TypeMeta,
		Namespace: manifest.Metadata.Namespace,
		Name:      manifest.Metadata.Name,
	}
	if err := target.Validate(request.Scope); err != nil {
		return invalidDraftResult(), nil
	}
	planner := service.planners[manifest.TypeMeta]
	if planner == nil {
		return invalidDraftResult(), nil
	}
	typedSpec, err := service.registry.DecodeAndValidate(manifest)
	if err != nil {
		return invalidDraftResult(), nil
	}
	canonical, err := control.CanonicalJSONObject(manifest)
	if err != nil {
		return invalidDraftResult(), nil
	}
	result := ValidationResult{
		Target: target, CanonicalManifest: json.RawMessage(canonical),
		Issues: []control.PlanIssue{},
	}
	head, err := service.repository.GetResource(ctx, request.Scope, target)
	switch {
	case errors.Is(err, control.ErrNotFound):
		result.Operation = control.PlanOperationCreate
		if err := service.authorizer.AuthorizeCreate(ctx, request.Scope, target); err != nil {
			return validatedDraft{}, err
		}
	case err != nil:
		return validatedDraft{}, err
	default:
		result.Operation = control.PlanOperationUpdate
		if err := service.authorizer.AuthorizeUpdate(ctx, request.Scope, head); err != nil {
			return validatedDraft{}, err
		}
	}
	var headPointer *control.ResourceHead
	if result.Operation == control.PlanOperationUpdate {
		headPointer = &head
	}
	return validatedDraft{
		scope: request.Scope, result: result, manifest: manifest, typedSpec: typedSpec,
		planner: planner, head: headPointer,
	}, nil
}

func decodeResourceSource(source ResourceSource) (orchestrationresource.Manifest, error) {
	switch source.Format {
	case SourceFormatJSON:
		return orchestrationresource.DecodeJSONSubmission(source.Content)
	case SourceFormatYAML:
		return orchestrationresource.DecodeYAMLSubmission(source.Content)
	default:
		return orchestrationresource.Manifest{}, control.ErrInvalid
	}
}

func invalidDraftResult() validatedDraft {
	return validatedDraft{result: ValidationResult{
		Issues: []control.PlanIssue{{
			Severity: control.PlanIssueBlocking,
			Path:     "/", Code: "invalid-draft",
			Message: "The resource draft is invalid.",
		}},
	}}
}
