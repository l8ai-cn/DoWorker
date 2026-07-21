package agentpod

import (
	"context"
	"errors"
	"fmt"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/repository"
)

var ErrCreateResourceUnavailable = errors.New("create resource unavailable")

func (o *PodOrchestrator) preResolveFreshRepository(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	agentDef *agentDomain.Agent,
) error {
	if req.preResolvedDependencies != nil {
		return nil
	}
	if req.preparedWorkerSpec == nil || req.preResolvedRepository == nil {
		req.preResolvedRepository = nil
		req.preResolvedRepositorySlug = ""
	}

	layerRepo := repositoryDeclaration{}
	if req.AgentfileLayer != nil {
		var err error
		layerRepo, err = parseRepositoryDeclaration(*req.AgentfileLayer)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidAgentfileLayer, err)
		}
	}
	repo := layerRepo
	if !layerRepo.present && agentDef != nil && agentDef.AgentfileSource != nil {
		baseRepo, err := parseRepositoryDeclaration(*agentDef.AgentfileSource)
		if err != nil {
			return fmt.Errorf("base agentfile parse error: %v", err)
		}
		repo = baseRepo
	}
	if repo.slug != "" {
		return o.resolveAgentfileRepository(ctx, req, &agentfileResolved{}, repo.slug)
	}
	if req.RepositoryID == nil {
		return nil
	}
	return o.resolveEffectiveRepository(ctx, req, &agentfileResolved{}, req.RepositoryID)
}

func (o *PodOrchestrator) resolveAgentfileRepository(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
	slug string,
) error {
	if req.preResolvedRepository != nil && req.preResolvedRepositorySlug == slug {
		resolved.RepositoryID = &req.preResolvedRepository.ID
		resolved.Repository = req.preResolvedRepository
		return nil
	}
	if o.repoService == nil {
		return ErrCreateResourceUnavailable
	}
	repo, err := o.repoService.FindAccessibleByOrgSlug(ctx, req.OrganizationID, req.UserID, slug)
	if err != nil {
		return createRepositoryError(err)
	}
	if repo == nil {
		return ErrCreateResourceUnavailable
	}
	req.preResolvedRepository = repo
	req.preResolvedRepositorySlug = slug
	resolved.RepositoryID = &repo.ID
	resolved.Repository = repo
	return nil
}

func (o *PodOrchestrator) resolveEffectiveRepository(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
	repositoryID *int64,
) error {
	if req.preResolvedDependencies != nil {
		return nil
	}
	if repositoryID == nil {
		return nil
	}
	if req.preResolvedRepository != nil && req.preResolvedRepository.ID == *repositoryID {
		resolved.Repository = req.preResolvedRepository
		return nil
	}
	if resolved.Repository != nil && resolved.Repository.ID == *repositoryID {
		return nil
	}
	if o.repoService == nil {
		return ErrCreateResourceUnavailable
	}
	repo, err := o.repoService.GetAccessibleByID(ctx, *repositoryID, req.OrganizationID, req.UserID)
	if err != nil {
		return createRepositoryError(err)
	}
	if repo == nil {
		return ErrCreateResourceUnavailable
	}
	req.preResolvedRepository = repo
	req.preResolvedRepositorySlug = ""
	resolved.Repository = repo
	return nil
}

func createRepositoryError(err error) error {
	if errors.Is(err, repository.ErrNoPermission) ||
		errors.Is(err, repository.ErrRepositoryNotFound) ||
		errors.Is(err, repository.ErrAmbiguousRepositorySlug) {
		return ErrCreateResourceUnavailable
	}
	return err
}
