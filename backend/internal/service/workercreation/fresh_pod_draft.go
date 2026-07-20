package workercreation

import (
	"context"
	"fmt"
	"maps"
	"strings"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type FreshPodDraftInput struct {
	OptionsRevision        string
	OrganizationSlug       string
	WorkerTypeSlug         string
	ModelResourceID        *int64
	SkillIDs               []int64
	ToolModelResourceIDs   map[string]int64
	ConfigDocumentBindings []specdomain.ConfigDocumentBinding
	Runtime                specservice.RuntimeSelection
	RepositoryID           *int64
	Branch                 string
	Alias                  string
	AgentfileLayer         string
	AutomationLevel        specdomain.AutomationLevel
	Perpetual              bool
}

func (service *Service) NewFreshPodDraft(
	ctx context.Context,
	scope specservice.Scope,
	input FreshPodDraftInput,
) (Draft, error) {
	if service == nil || service.revision == "" || service.runtime == nil ||
		service.workerTypes == nil {
		return Draft{}, specservice.ErrResolverUnavailable
	}
	if scope.OrgID <= 0 || scope.UserID <= 0 {
		return Draft{}, specservice.ErrInvalidScope
	}
	if strings.TrimSpace(input.OptionsRevision) != service.revision {
		return Draft{}, ErrStaleOptions
	}
	namespace, err := slugkit.NewFromTrusted(strings.TrimSpace(input.OrganizationSlug))
	if err != nil {
		return Draft{}, invalidFreshPodDraft("organization_slug", err.Error())
	}
	workerType, err := slugkit.NewFromTrusted(strings.TrimSpace(input.WorkerTypeSlug))
	if err != nil {
		return Draft{}, invalidFreshPodDraft("worker_type_slug", err.Error())
	}
	resolution, err := service.workerTypes.ResolveWorkerType(ctx, scope, workerType)
	if err != nil {
		return Draft{}, err
	}
	if resolution.WorkerType.Slug != workerType {
		return Draft{}, invalidFreshPodDraft(
			"worker_type_slug",
			"worker type resolver substituted the requested slug",
		)
	}
	if err := service.validateFreshPodRuntime(ctx, scope, workerType, input.Runtime); err != nil {
		return Draft{}, err
	}
	if err := validateFreshPodAutomation(input.AutomationLevel); err != nil {
		return Draft{}, err
	}
	layer, err := parseFreshPodAgentfileLayer(input.AgentfileLayer)
	if err != nil {
		return Draft{}, err
	}
	secretRefs, err := service.freshPodSecretRefs(ctx, scope, workerType, resolution.TypeSchema)
	if err != nil {
		return Draft{}, err
	}
	return Draft{
		OptionsRevision:  service.revision,
		OrganizationSlug: namespace,
		WorkerSpec: specservice.Draft{
			ModelResourceID:      modelResourceIDValue(input.ModelResourceID),
			ToolModelResourceIDs: maps.Clone(input.ToolModelResourceIDs),
			WorkerTypeSlug:       workerType,
			Runtime:              input.Runtime,
			TypeConfig: specdomain.TypeConfig{
				SchemaVersion:   resolution.TypeSchema.Version,
				Values:          layer.config,
				SecretRefs:      secretRefs,
				InteractionMode: layer.interactionMode,
				AutomationLevel: input.AutomationLevel,
			},
			Workspace: specdomain.Workspace{
				RepositoryID: input.RepositoryID,
				Branch:       firstNonEmpty(input.Branch, layer.branch),
				SkillIDs:     append([]int64{}, input.SkillIDs...),
				ConfigDocumentBindings: append(
					[]specdomain.ConfigDocumentBinding{},
					input.ConfigDocumentBindings...,
				),
				InitialTask: layer.prompt,
			},
			Lifecycle: freshPodLifecycle(input.Perpetual),
			Metadata:  specdomain.Metadata{Alias: strings.TrimSpace(input.Alias)},
		},
	}, nil
}

func validateFreshPodAutomation(level specdomain.AutomationLevel) error {
	switch level {
	case specdomain.AutomationLevelInteractive,
		specdomain.AutomationLevelAutoEdit,
		specdomain.AutomationLevelAutonomous:
		return nil
	default:
		return invalidFreshPodDraft("automation_level", "unsupported value")
	}
}

func (service *Service) validateFreshPodRuntime(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	selection specservice.RuntimeSelection,
) error {
	if selection.RuntimeImageID <= 0 {
		return invalidFreshPodDraft("runtime_image_id", "must be positive")
	}
	if selection.ComputeTargetID <= 0 {
		return invalidFreshPodDraft("compute_target_id", "must be positive")
	}
	if selection.ResourceProfileID <= 0 && selection.CustomResources == nil {
		return invalidFreshPodDraft("resource_profile_id", "must be positive")
	}
	resolved, err := service.runtime.ResolveRuntime(
		ctx,
		scope,
		workerType,
		selection,
	)
	if err != nil {
		return err
	}
	return specservice.ValidateRuntimeSelection(selection, resolved)
}

func modelResourceIDValue(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func freshPodLifecycle(perpetual bool) specdomain.Lifecycle {
	if perpetual {
		return specdomain.Lifecycle{
			TerminationPolicy: specdomain.TerminationPolicyManual,
		}
	}
	return specdomain.Lifecycle{
		TerminationPolicy:  specdomain.TerminationPolicyOnIdle,
		IdleTimeoutMinutes: 30,
	}
}

func invalidFreshPodDraft(field, reason string) error {
	return &specservice.InvalidDraftFieldError{
		Field:  field,
		Reason: fmt.Sprintf("fresh pod draft: %s", reason),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
