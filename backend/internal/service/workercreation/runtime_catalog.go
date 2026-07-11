package workercreation

import (
	"context"
	"fmt"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type runtimeCatalogResolver struct {
	catalog runtimedomain.Catalog
}

func newRuntimeCatalogResolver(catalog runtimedomain.Catalog) *runtimeCatalogResolver {
	return &runtimeCatalogResolver{catalog: catalog}
}

func (resolver *runtimeCatalogResolver) ResolveRuntime(
	_ context.Context,
	_ specservice.Scope,
	workerType slugkit.Slug,
	selection specservice.RuntimeSelection,
) (runtimedomain.Resolved, error) {
	image, err := resolver.resolveImage(workerType.String(), selection.RuntimeImageID)
	if err != nil {
		return runtimedomain.Resolved{}, err
	}
	target := resolver.catalog.Target(selection.ComputeTargetID)
	if target == nil {
		return runtimedomain.Resolved{}, invalidRuntimeSelection(
			"compute target",
			"selection does not exist",
		)
	}
	if !target.Enabled {
		reason := target.DisabledReason
		if reason == "" {
			reason = "selection is disabled"
		}
		return runtimedomain.Resolved{}, invalidRuntimeSelection("compute target", reason)
	}
	if !supportsDeploymentMode(*target, selection.DeploymentMode) {
		return runtimedomain.Resolved{}, invalidRuntimeSelection(
			"deployment mode",
			fmt.Sprintf("%q is not supported by compute target %q", selection.DeploymentMode, target.Slug),
		)
	}
	profile := resolver.catalog.Profile(selection.ResourceProfileID)
	if profile == nil {
		return runtimedomain.Resolved{}, invalidRuntimeSelection(
			"resource profile",
			"selection does not exist",
		)
	}
	if !profile.Enabled {
		return runtimedomain.Resolved{}, invalidRuntimeSelection(
			"resource profile",
			"selection is disabled",
		)
	}
	runtimeImage, placement, err := specdomain.NormalizeAndValidateRuntimePlacement(
		specdomain.RuntimeImage{ID: image.ID, Digest: image.Digest},
		specdomain.Placement{
			Policy: selection.PlacementPolicy,
			ComputeTarget: specdomain.ComputeTarget{
				ID:   target.ID,
				Kind: target.Kind,
			},
			DeploymentMode: selection.DeploymentMode,
			ResourceProfile: specdomain.ResourceProfile{
				ID:        profile.ID,
				Resources: profile.Resources,
			},
		},
	)
	if err != nil {
		return runtimedomain.Resolved{}, invalidRuntimeSelection("runtime", err.Error())
	}
	return runtimedomain.Resolved{
		RuntimeImage: runtimeImage,
		Placement:    placement,
	}, nil
}

func (resolver *runtimeCatalogResolver) resolveImage(
	workerType string,
	imageID int64,
) (runtimedomain.CatalogRuntimeImage, error) {
	for _, image := range resolver.catalog.ImagesFor(workerType) {
		if image.ID != imageID {
			continue
		}
		if !image.Enabled {
			return runtimedomain.CatalogRuntimeImage{}, invalidRuntimeSelection(
				"runtime image",
				"selection is disabled",
			)
		}
		return image, nil
	}
	return runtimedomain.CatalogRuntimeImage{}, invalidRuntimeSelection(
		"runtime image",
		"selection is not available for the worker type",
	)
}

func supportsDeploymentMode(
	target runtimedomain.CatalogComputeTarget,
	mode specdomain.DeploymentMode,
) bool {
	switch mode {
	case specdomain.DeploymentModePooled:
		return target.SupportsPooled
	case specdomain.DeploymentModeDedicated:
		return target.SupportsDedicated
	default:
		return false
	}
}

func invalidRuntimeSelection(field, reason string) error {
	return fmt.Errorf("%w: %s: %s", specservice.ErrInvalidDraft, field, reason)
}
