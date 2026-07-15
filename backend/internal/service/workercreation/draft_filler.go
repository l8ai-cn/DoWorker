package workercreation

import (
	"context"
	"fmt"
	"strings"

	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type DraftJSONGenerator interface {
	Generate(
		context.Context,
		*resourceservice.ResolvedResource,
		string,
		string,
	) ([]byte, error)
}

type DraftFiller struct {
	service   *Service
	resources ModelResourceResolver
	generator DraftJSONGenerator
}

func NewDraftFiller(
	service *Service,
	resources ModelResourceResolver,
	generator DraftJSONGenerator,
) *DraftFiller {
	return &DraftFiller{
		service:   service,
		resources: resources,
		generator: generator,
	}
}

func (filler *DraftFiller) Fill(
	ctx context.Context,
	scope specservice.Scope,
	prompt string,
	generationModelResourceID int64,
	current *Draft,
) (FillResult, error) {
	if filler == nil || filler.service == nil || filler.resources == nil ||
		filler.generator == nil || filler.service.workerTypes == nil {
		return FillResult{}, specservice.ErrResolverUnavailable
	}
	if scope.OrgID <= 0 || scope.UserID <= 0 {
		return FillResult{}, specservice.ErrInvalidScope
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return FillResult{}, invalidFillField("prompt", "is required")
	}
	if current == nil {
		return FillResult{}, invalidFillField("current_draft", "is required")
	}
	if current.OptionsRevision != filler.service.Revision() {
		return FillResult{}, ErrStaleOptions
	}
	workerType, err := filler.service.workerTypes.ResolveWorkerType(
		ctx,
		scope,
		current.WorkerSpec.WorkerTypeSlug,
	)
	if err != nil {
		return FillResult{}, err
	}
	resource, err := filler.resolveModelResource(
		ctx,
		scope,
		generationModelResourceID,
	)
	if err != nil {
		return FillResult{}, err
	}
	systemPrompt, userPrompt, err := buildDraftFillPrompts(
		prompt,
		*current,
		workerType,
	)
	if err != nil {
		return FillResult{}, err
	}
	rawPatch, err := filler.generator.Generate(
		ctx,
		resource,
		systemPrompt,
		userPrompt,
	)
	if err != nil {
		return FillResult{}, fmt.Errorf("generate worker draft patch: %w", err)
	}
	patch, err := decodeDraftFillPatch(rawPatch)
	if err != nil {
		return FillResult{}, err
	}
	draft, err := applyDraftFillPatch(*current, patch)
	if err != nil {
		return FillResult{}, err
	}
	preflight, err := filler.service.Preflight(ctx, scope, draft)
	if err != nil {
		return FillResult{}, err
	}
	issues := append([]Issue{}, preflight.BlockingErrors...)
	issues = append(issues, preflight.Warnings...)
	return FillResult{Draft: draft, Issues: issues}, nil
}

func (filler *DraftFiller) resolveModelResource(
	ctx context.Context,
	scope specservice.Scope,
	resourceID int64,
) (*resourceservice.ResolvedResource, error) {
	if resourceID <= 0 {
		return nil, invalidFillField(
			"generation_model_resource_id",
			"must be positive",
		)
	}
	requirements := draftGenerationModelRequirements()
	resolved, err := filler.resources.ResolveExact(
		ctx,
		resourceservice.Actor{UserID: scope.UserID},
		scope.OrgID,
		resourceID,
		requirements,
	)
	if err != nil {
		if isModelSelectionError(err) {
			return nil, fmt.Errorf(
				"%w: model resource: %w",
				specservice.ErrInvalidDraft,
				err,
			)
		}
		return nil, err
	}
	if err := validateResolvedModel(resolved, resourceID); err != nil {
		return nil, err
	}
	if resolved.Provider.Key != resolved.Connection.ProviderKey {
		return nil, invalidResolvedModel("provider definition does not match connection")
	}
	if !containsString(
		requirements.AllowedProtocolAdapters,
		resolved.Provider.ProtocolAdapter,
	) {
		return nil, invalidResolvedModel("provider protocol was substituted")
	}
	return resolved, nil
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func invalidFillField(field, reason string) error {
	return &specservice.InvalidDraftFieldError{Field: field, Reason: reason}
}
