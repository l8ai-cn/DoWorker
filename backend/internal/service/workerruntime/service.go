package workerruntime

import (
	"context"
	"fmt"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type Service struct {
	repository domain.Repository
}

func NewService(repository domain.Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Resolve(
	ctx context.Context,
	request domain.Request,
) (domain.Resolved, error) {
	if err := domain.ValidateRequest(request); err != nil {
		return domain.Resolved{}, err
	}
	if service == nil || service.repository == nil {
		return domain.Resolved{}, domain.ErrRepositoryUnavailable
	}
	image, err := service.repository.GetRuntimeImageByIDForOrganization(
		ctx,
		request.OrganizationID,
		request.RuntimeImageID,
	)
	if err != nil {
		return domain.Resolved{}, err
	}
	if err := validateImage(image, request.RuntimeImageID); err != nil {
		return domain.Resolved{}, err
	}
	target, err := service.repository.GetComputeTargetByIDForOrganization(
		ctx,
		request.OrganizationID,
		request.ComputeTargetID,
	)
	if err != nil {
		return domain.Resolved{}, err
	}
	if err := validateTarget(target, request.ComputeTargetID); err != nil {
		return domain.Resolved{}, err
	}
	profile, err := service.repository.GetResourceProfileByIDForOrganization(
		ctx,
		request.OrganizationID,
		request.ResourceProfileID,
	)
	if err != nil {
		return domain.Resolved{}, err
	}
	if err := validateProfile(profile, request.ResourceProfileID); err != nil {
		return domain.Resolved{}, err
	}
	if err := service.validateCompatibility(ctx, request); err != nil {
		return domain.Resolved{}, err
	}
	imageValue, placement, err := workerspec.NormalizeAndValidateRuntimePlacement(
		workerspec.RuntimeImage{ID: image.ID, Digest: image.Digest},
		workerspec.Placement{
			Policy: request.PlacementPolicy,
			ComputeTarget: workerspec.ComputeTarget{
				ID:   target.ID,
				Kind: target.Kind,
			},
			DeploymentMode: request.DeploymentMode,
			ResourceProfile: workerspec.ResourceProfile{
				ID:        profile.ID,
				Resources: profile.Resources,
			},
		},
	)
	if err != nil {
		return domain.Resolved{}, fmt.Errorf("%w: %v", domain.ErrInvalidResolvedValue, err)
	}
	return domain.Resolved{
		RuntimeImage: imageValue,
		Placement:    placement,
	}, nil
}

func (service *Service) validateCompatibility(
	ctx context.Context,
	request domain.Request,
) error {
	imageCompatible, err := service.repository.IsRuntimeImageCompatibleWithWorkerType(
		ctx,
		request.OrganizationID,
		request.WorkerTypeSlug,
		request.RuntimeImageID,
	)
	if err != nil {
		return err
	}
	if !imageCompatible {
		return domain.ErrIncompatible
	}
	targetCompatible, err := service.repository.IsComputeTargetCompatibleWithDeployment(
		ctx,
		request.OrganizationID,
		request.ComputeTargetID,
		request.DeploymentMode,
	)
	if err != nil {
		return err
	}
	if !targetCompatible {
		return domain.ErrIncompatible
	}
	profileCompatible, err := service.repository.IsComputeTargetCompatibleWithResourceProfile(
		ctx,
		request.OrganizationID,
		request.ComputeTargetID,
		request.ResourceProfileID,
	)
	if err != nil {
		return err
	}
	if !profileCompatible {
		return domain.ErrIncompatible
	}
	return nil
}

func validateImage(image *domain.RuntimeImage, expectedID int64) error {
	if image == nil {
		return domain.ErrNotFound
	}
	if image.ID != expectedID {
		return domain.ErrInvalidResolvedValue
	}
	if !image.Enabled {
		return domain.ErrDisabled
	}
	return nil
}

func validateTarget(target *domain.ComputeTarget, expectedID int64) error {
	if target == nil {
		return domain.ErrNotFound
	}
	if target.ID != expectedID {
		return domain.ErrInvalidResolvedValue
	}
	if !target.Enabled {
		return domain.ErrDisabled
	}
	return nil
}

func validateProfile(profile *domain.ResourceProfile, expectedID int64) error {
	if profile == nil {
		return domain.ErrNotFound
	}
	if profile.ID != expectedID {
		return domain.ErrInvalidResolvedValue
	}
	if !profile.Enabled {
		return domain.ErrDisabled
	}
	return nil
}
