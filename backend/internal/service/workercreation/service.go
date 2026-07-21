package workercreation

import (
	"context"
	"errors"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	runtimedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

type Service struct {
	revision       string
	catalog        runtimedomain.Catalog
	modelResources ModelResourceResolver
	agents         AgentProvider
	workerTypes    specservice.WorkerTypeResolver
	runtime        specservice.RuntimeResolver
	models         specservice.ModelResolver
	toolModels     specservice.ToolModelResolver
	runners        RunnerAvailabilityResolver
	workspaceDeps  workspaceResolverDeps
}

func NewService(deps Deps) *Service {
	workspaceDeps := workspaceResolverDeps{
		Repositories: deps.Repositories,
		Skills:       deps.Skills,
		Knowledge:    deps.Knowledge,
		EnvBundles:   deps.EnvBundles,
		Definitions:  deps.Definitions,
		Commits:      deps.Commits,
	}
	workerTypes := newWorkerTypeResolver(deps.Agents, deps.Definitions)
	models := newModelResolver(deps.Models)
	return &Service{
		revision:       deps.Catalog.Revision(),
		catalog:        deps.Catalog,
		modelResources: deps.Models,
		agents:         deps.Agents,
		workerTypes:    workerTypes,
		runtime:        newRuntimeCatalogResolver(deps.Catalog),
		models:         models,
		toolModels:     models,
		runners:        deps.Runners,
		workspaceDeps:  workspaceDeps,
	}
}

func (service *Service) Revision() string {
	if service == nil {
		return ""
	}
	return service.revision
}

func (service *Service) Prepare(
	ctx context.Context,
	scope specservice.Scope,
	draft Draft,
) (Prepared, error) {
	if service == nil || service.workerTypes == nil || service.runtime == nil ||
		service.modelResources == nil || service.revision == "" {
		return Prepared{}, specservice.ErrResolverUnavailable
	}
	if draft.OptionsRevision != service.revision {
		return Prepared{}, ErrStaleOptions
	}
	workspace := newWorkspaceResolver(service.workspaceDeps)
	runtime := newRuntimeCatalogResolver(service.catalog)
	models := newModelResolver(service.modelResources)
	resolver := specservice.NewResolver(specservice.ResolverDeps{
		WorkerTypes: service.workerTypes,
		Runtime:     runtime,
		Models:      models,
		ToolModels:  models,
		Secrets:     workspace,
		Workspaces:  workspace,
	})
	resolved, err := resolver.Resolve(ctx, scope, draft.WorkerSpec)
	if err != nil {
		return Prepared{}, err
	}
	spec, err := specdomain.DecodeSpec(resolved.SpecJSON())
	if err != nil {
		return Prepared{}, fmt.Errorf("decode prepared workerspec: %w", err)
	}
	layer, err := newCompiler(workspace).Compile(ctx, scope, spec)
	if err != nil {
		return Prepared{}, err
	}
	artifact, err := service.buildArtifact(
		ctx,
		scope,
		draft.OrganizationSlug,
		draft.ArtifactRefs,
		spec,
		layer,
		models,
		runtime,
		workspace,
	)
	if err != nil {
		return Prepared{}, err
	}
	if artifact == nil {
		return Prepared{}, fmt.Errorf("worker dependency artifact is required")
	}
	document, err := workerdependency.Decode(artifact.JSON())
	if err != nil {
		return Prepared{}, fmt.Errorf("decode worker dependency artifact: %w", err)
	}
	return Prepared{
		Snapshot:       resolved,
		Spec:           spec,
		AgentfileLayer: layer,
		Repository:     workspace.resolvedRepository(spec.Workspace.RepositoryID),
		Artifact:       artifact,
		Dependencies:   &document,
	}, nil
}

func (service *Service) Preflight(
	ctx context.Context,
	scope specservice.Scope,
	draft Draft,
) (PreflightResult, error) {
	result := PreflightResult{
		BlockingErrors:  []Issue{},
		Warnings:        []Issue{},
		OptionsRevision: service.Revision(),
	}
	prepared, err := service.Prepare(ctx, scope, draft)
	if err == nil {
		result.Resolved = &prepared
		return result, nil
	}
	switch {
	case errors.Is(err, ErrStaleOptions):
		result.BlockingErrors = append(result.BlockingErrors, Issue{
			Code:     "stale-options",
			Field:    "options_revision",
			Message:  err.Error(),
			Severity: "blocking",
		})
		return result, nil
	case errors.Is(err, specservice.ErrInvalidDraft):
		field := specservice.InvalidDraftField(err)
		if field == "" {
			field = "draft"
		} else {
			field = "worker_spec." + field
		}
		result.BlockingErrors = append(result.BlockingErrors, Issue{
			Code:     "invalid-draft",
			Field:    field,
			Message:  err.Error(),
			Severity: "blocking",
		})
		return result, nil
	default:
		return result, err
	}
}
