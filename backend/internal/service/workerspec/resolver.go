package workerspec

import (
	"context"
	"fmt"
	"sort"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Resolver struct {
	workerTypes WorkerTypeResolver
	runtime     RuntimeResolver
	models      ModelResolver
	toolModels  ToolModelResolver
	secrets     SecretReferenceResolver
	workspaces  WorkspaceResolver
}

func NewResolver(deps ResolverDeps) *Resolver {
	return &Resolver{
		workerTypes: deps.WorkerTypes,
		runtime:     deps.Runtime,
		models:      deps.Models,
		toolModels:  deps.ToolModels,
		secrets:     deps.Secrets,
		workspaces:  deps.Workspaces,
	}
}

func (resolver *Resolver) Resolve(
	ctx context.Context,
	scope Scope,
	draft Draft,
) (ResolvedSnapshot, error) {
	if err := validateScope(scope); err != nil {
		return ResolvedSnapshot{}, err
	}
	if err := resolver.validateDependencies(); err != nil {
		return ResolvedSnapshot{}, err
	}
	workerType, err := resolver.workerTypes.ResolveWorkerType(
		ctx,
		scope,
		draft.WorkerTypeSlug,
	)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("resolve worker type: %w", err)
	}
	if err := validateWorkerTypeResolution(draft.WorkerTypeSlug, workerType); err != nil {
		return ResolvedSnapshot{}, err
	}
	if err := validateInteractionMode(
		draft.TypeConfig.InteractionMode,
		workerType.SupportedInteractionModes,
	); err != nil {
		return ResolvedSnapshot{}, err
	}
	runtime, err := resolver.runtime.ResolveRuntime(
		ctx,
		scope,
		workerType.WorkerType.Slug,
		draft.Runtime,
	)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("resolve worker runtime: %w", err)
	}
	if err := validateRuntimeResolution(draft.Runtime, runtime); err != nil {
		return ResolvedSnapshot{}, err
	}
	modelBinding, err := resolver.resolveModelBinding(
		ctx,
		scope,
		workerType.ModelRequirement,
		draft.ModelResourceID,
	)
	if err != nil {
		return ResolvedSnapshot{}, err
	}
	toolModelBindings, err := resolver.resolveToolModelBindings(
		ctx,
		scope,
		workerType.ToolModelRequirements,
		draft.ToolModelResourceIDs,
	)
	if err != nil {
		return ResolvedSnapshot{}, err
	}
	if err := domain.ValidateTypeConfigAgainstSchema(
		draft.TypeConfig,
		workerType.TypeSchema,
	); err != nil {
		return ResolvedSnapshot{}, fmt.Errorf(
			"%w: validate worker type config: %v",
			ErrInvalidDraft,
			err,
		)
	}
	if err := resolver.resolveSecretReferences(
		ctx,
		scope,
		workerType.WorkerType.Slug,
		draft.TypeConfig.SecretRefs,
	); err != nil {
		return ResolvedSnapshot{}, err
	}
	workspace, err := resolver.workspaces.ResolveWorkspace(
		ctx,
		scope,
		workerType.WorkerType.Slug,
		cloneWorkspace(draft.Workspace),
	)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("resolve worker workspace: %w", err)
	}
	spec, err := domain.NormalizeAndValidate(domain.NewV1(
		domain.Runtime{
			ModelBinding:      modelBinding,
			ToolModelBindings: toolModelBindings,
			WorkerType:        workerType.WorkerType,
			Image:             runtime.RuntimeImage,
		},
		runtime.Placement,
		draft.TypeConfig,
		workspace,
		draft.Lifecycle,
		draft.Metadata,
	))
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf(
			"%w: validate resolved workerspec: %v",
			ErrInvalidDraft,
			err,
		)
	}
	return resolveSnapshot(scope.OrgID, spec)
}

func (resolver *Resolver) validateDependencies() error {
	if resolver == nil ||
		resolver.workerTypes == nil ||
		resolver.runtime == nil ||
		resolver.secrets == nil ||
		resolver.workspaces == nil {
		return ErrResolverUnavailable
	}
	return nil
}

func validateInteractionMode(
	requested domain.InteractionMode,
	supported []domain.InteractionMode,
) error {
	for _, mode := range supported {
		if mode == requested {
			return nil
		}
	}
	return &InvalidDraftFieldError{
		Field: "type_config.interaction_mode",
		Reason: fmt.Sprintf(
			"interaction mode %q is not supported by the selected worker type",
			requested,
		),
	}
}

func (resolver *Resolver) resolveSecretReferences(
	ctx context.Context,
	scope Scope,
	workerType slugkit.Slug,
	references map[string]domain.SecretReference,
) error {
	fields := make([]string, 0, len(references))
	for field := range references {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if err := resolver.secrets.ResolveSecretReference(
			ctx,
			scope,
			workerType,
			field,
			references[field],
		); err != nil {
			return fmt.Errorf("resolve secret reference %q: %w", field, err)
		}
	}
	return nil
}
