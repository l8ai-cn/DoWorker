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
	selection, err := service.repository.ResolveSelection(ctx, request)
	if err != nil {
		return domain.Resolved{}, err
	}
	if selection == nil {
		return domain.Resolved{}, domain.ErrNotFound
	}
	if err := validateImage(selection.RuntimeImage, request.RuntimeImageID); err != nil {
		return domain.Resolved{}, err
	}
	if err := validateTarget(selection.ComputeTarget, request.ComputeTargetID); err != nil {
		return domain.Resolved{}, err
	}
	if err := validateProfile(selection.ResourceProfile, request.ResourceProfileID); err != nil {
		return domain.Resolved{}, err
	}
	if !selection.ImageCompatible ||
		!selection.DeploymentCompatible ||
		!selection.ResourceProfileCompatible {
		return domain.Resolved{}, domain.ErrIncompatible
	}
	image := selection.RuntimeImage
	target := selection.ComputeTarget
	profile := selection.ResourceProfile
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
