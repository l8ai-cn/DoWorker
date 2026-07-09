package agentpod

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
)

var ErrCreateResourceUnavailable = errors.New("create resource unavailable")

func (o *PodOrchestrator) preResolveFreshRepository(ctx context.Context, req *OrchestrateCreatePodRequest) error {
	req.preResolvedRepository = nil
	req.preResolvedRepositorySlug = ""

	repoSlug := ""
	if req.AgentfileLayer != nil {
		repoSlug = peekRepoSlug(*req.AgentfileLayer)
	}
	if repoSlug == "" && o.agentResolver != nil {
		agentDef, err := o.agentResolver.GetAgent(ctx, req.AgentSlug)
		if err == nil && agentDef != nil && agentDef.AgentfileSource != nil {
			repoSlug = peekRepoSlug(*agentDef.AgentfileSource)
		}
	}
	if repoSlug != "" {
		return o.resolveAgentfileRepository(ctx, req, &agentfileResolved{}, repoSlug)
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
