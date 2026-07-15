package workercreation

import (
	"context"
	"errors"
	"fmt"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type OptionsFilter struct {
	WorkerTypeSlug  string
	ComputeTargetID *int64
	DeploymentMode  specdomain.DeploymentMode
}

type CreateOptions struct {
	Revision         string
	WorkerTypes      []WorkerTypeOption
	RuntimeImages    []RuntimeImageOption
	ComputeTargets   []ComputeTargetOption
	DeploymentModes  []DeploymentModeOption
	ResourceProfiles []ResourceProfileOption
}

type WorkerTypeOption struct {
	Slug                      string
	Name                      string
	Description               string
	Schema                    specdomain.TypeSchema
	SupportedInteractionModes []specdomain.InteractionMode
	RequiresModelResource     bool
	ToolModelRequirements     []specdomain.ToolModelRequirement
	Selectable                bool
	BlockingReason            string
}

type RuntimeImageOption struct {
	Image          runtimedomain.CatalogRuntimeImage
	Selectable     bool
	BlockingReason string
}

type ComputeTargetOption struct {
	Target         runtimedomain.CatalogComputeTarget
	Selectable     bool
	BlockingReason string
}

type DeploymentModeOption struct {
	Value          specdomain.DeploymentMode
	Name           string
	Selectable     bool
	BlockingReason string
}

type ResourceProfileOption struct {
	Profile        runtimedomain.CatalogResourceProfile
	Selectable     bool
	BlockingReason string
}

func (service *Service) ListOptions(
	ctx context.Context,
	scope specservice.Scope,
	filter OptionsFilter,
) (CreateOptions, error) {
	if service == nil || service.agents == nil || service.workerTypes == nil ||
		service.runners == nil || service.revision == "" {
		return CreateOptions{}, specservice.ErrResolverUnavailable
	}
	if scope.OrgID <= 0 || scope.UserID <= 0 {
		return CreateOptions{}, specservice.ErrInvalidScope
	}
	if filter.DeploymentMode != "" &&
		filter.DeploymentMode != specdomain.DeploymentModePooled &&
		filter.DeploymentMode != specdomain.DeploymentModeDedicated {
		return CreateOptions{}, &specservice.InvalidDraftFieldError{
			Field:  "deployment_mode",
			Reason: fmt.Sprintf("unsupported value %q", filter.DeploymentMode),
		}
	}
	workerTypes, err := service.listWorkerTypeOptions(ctx, scope, filter.WorkerTypeSlug)
	if err != nil {
		return CreateOptions{}, err
	}
	return CreateOptions{
		Revision:         service.revision,
		WorkerTypes:      workerTypes,
		RuntimeImages:    runtimeImageOptions(service.catalog, filter.WorkerTypeSlug),
		ComputeTargets:   computeTargetOptions(service.catalog),
		DeploymentModes:  deploymentModeOptions(service.catalog, filter.ComputeTargetID),
		ResourceProfiles: resourceProfileOptions(service.catalog),
	}, nil
}

func (service *Service) listWorkerTypeOptions(
	ctx context.Context,
	scope specservice.Scope,
	filter string,
) ([]WorkerTypeOption, error) {
	agents, err := service.agents.ListBuiltinAgents(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]WorkerTypeOption, 0, len(agents))
	for _, agent := range agents {
		if agent == nil || !agent.IsActive || agent.IsInternal ||
			(filter != "" && agent.Slug != filter) {
			continue
		}
		slug, err := slugkit.NewFromTrusted(agent.Slug)
		if err != nil {
			return nil, fmt.Errorf("invalid worker type slug %q: %w", agent.Slug, err)
		}
		option := WorkerTypeOption{Slug: agent.Slug, Name: agent.Name}
		if agent.Description != nil {
			option.Description = *agent.Description
		}
		resolved, err := service.workerTypes.ResolveWorkerType(ctx, scope, slug)
		if err != nil {
			if !errors.Is(err, specservice.ErrInvalidDraft) {
				return nil, err
			}
			option.BlockingReason = err.Error()
			options = append(options, option)
			continue
		}
		option.Schema = resolved.TypeSchema
		option.SupportedInteractionModes = append(
			[]specdomain.InteractionMode{},
			resolved.SupportedInteractionModes...,
		)
		option.RequiresModelResource = resolved.ModelRequirement.Required
		option.ToolModelRequirements = append(
			[]specdomain.ToolModelRequirement{},
			resolved.ToolModelRequirements...,
		)
		if !hasEnabledRuntimeImage(service.catalog, agent.Slug) {
			option.BlockingReason = "No runtime image is available for this worker type"
			options = append(options, option)
			continue
		}
		available, err := service.runners.HasAvailableRunnerForAgent(
			ctx, scope.OrgID, scope.UserID, agent.Slug,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"check Runner availability for worker type %q: %w",
				agent.Slug, err,
			)
		}
		if !available {
			option.BlockingReason = "No online Runner currently supports this worker type"
			options = append(options, option)
			continue
		}
		option.Selectable = true
		options = append(options, option)
	}
	return options, nil
}

func hasEnabledRuntimeImage(catalog runtimedomain.Catalog, workerType string) bool {
	for _, image := range catalog.ImagesFor(workerType) {
		if image.Enabled {
			return true
		}
	}
	return false
}

func runtimeImageOptions(
	catalog runtimedomain.Catalog,
	workerType string,
) []RuntimeImageOption {
	images := catalog.Images()
	if workerType != "" {
		images = catalog.ImagesFor(workerType)
	}
	options := make([]RuntimeImageOption, 0, len(images))
	for _, image := range images {
		option := RuntimeImageOption{Image: image, Selectable: image.Enabled}
		if !image.Enabled {
			option.BlockingReason = "Runtime image is disabled"
		}
		options = append(options, option)
	}
	return options
}
