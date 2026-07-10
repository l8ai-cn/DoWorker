package workerspec

import (
	"context"
	"fmt"
	"sort"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type Resolver struct {
	workerTypes WorkerTypeResolver
	runtime     RuntimeResolver
	models      ModelResolver
	secrets     SecretReferenceResolver
	workspaces  WorkspaceResolver
}

func NewResolver(deps ResolverDeps) *Resolver {
	return &Resolver{
		workerTypes: deps.WorkerTypes,
		runtime:     deps.Runtime,
		models:      deps.Models,
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
	modelBinding, err := resolver.models.ResolveModel(
		ctx,
		scope,
		draft.ModelResourceID,
	)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("resolve worker model: %w", err)
	}
	if err := validateModelResolution(draft.ModelResourceID, modelBinding); err != nil {
		return ResolvedSnapshot{}, err
	}
	if err := domain.ValidateTypeConfigAgainstSchema(
		draft.TypeConfig,
		workerType.TypeSchema,
	); err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("validate worker type config: %w", err)
	}
	if err := resolver.resolveSecretReferences(ctx, scope, draft.TypeConfig.SecretRefs); err != nil {
		return ResolvedSnapshot{}, err
	}
	workspace, err := resolver.workspaces.ResolveWorkspace(
		ctx,
		scope,
		cloneWorkspace(draft.Workspace),
	)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("resolve worker workspace: %w", err)
	}
	spec, err := domain.NormalizeAndValidate(domain.NewV1(
		domain.Runtime{
			ModelBinding: modelBinding,
			WorkerType:   workerType.WorkerType,
			Image:        runtime.RuntimeImage,
		},
		runtime.Placement,
		draft.TypeConfig,
		workspace,
		draft.Lifecycle,
		draft.Metadata,
	))
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("validate resolved workerspec: %w", err)
	}
	return resolveSnapshot(scope.OrgID, spec)
}

func (resolver *Resolver) validateDependencies() error {
	if resolver == nil ||
		resolver.workerTypes == nil ||
		resolver.runtime == nil ||
		resolver.models == nil ||
		resolver.secrets == nil ||
		resolver.workspaces == nil {
		return ErrResolverUnavailable
	}
	return nil
}

func (resolver *Resolver) resolveSecretReferences(
	ctx context.Context,
	scope Scope,
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
			field,
			references[field],
		); err != nil {
			return fmt.Errorf("resolve secret reference %q: %w", field, err)
		}
	}
	return nil
}
