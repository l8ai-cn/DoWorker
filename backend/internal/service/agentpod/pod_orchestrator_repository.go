package agentpod

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
)

var ErrCreateResourceUnavailable = errors.New("create resource unavailable")

func (o *PodOrchestrator) resolveAgentfileRepository(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
	slug string,
) error {
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
