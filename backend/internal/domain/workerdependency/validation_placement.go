package workerdependency

import (
	"fmt"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/distribution/reference"
)

func validatePlacement(document Document, placement Placement) error {
	if err := requireNormalized(
		"worker dependency catalog revision",
		placement.CatalogRevision,
	); err != nil {
		return err
	}
	image := placement.RuntimeImage
	if image.ID <= 0 {
		return fmt.Errorf("worker dependency runtime image id must be positive")
	}
	if err := requireNormalized(
		"worker dependency runtime image reference",
		image.Reference,
	); err != nil {
		return err
	}
	if err := validateDigest("worker dependency runtime image digest", image.Digest); err != nil {
		return err
	}
	named, err := reference.ParseNormalizedNamed(image.Reference)
	if err != nil {
		return fmt.Errorf("worker dependency runtime image reference is invalid: %w", err)
	}
	canonical, ok := named.(reference.Canonical)
	if !ok || canonical.Digest().String() != image.Digest {
		return fmt.Errorf("worker dependency runtime image reference must match its digest")
	}
	if err := validatePin(
		document,
		placement.ComputeTarget,
		resource.KindComputeTarget,
	); err != nil {
		return err
	}
	if placement.ComputeTarget.DomainID != placement.Spec.ComputeTarget.ID {
		return fmt.Errorf("worker dependency compute target id does not match placement")
	}
	if placement.Spec.ResourceProfile.Custom {
		if placement.ResourceProfile != nil {
			return fmt.Errorf("custom resource profile must not have a resource pin")
		}
	} else {
		if placement.ResourceProfile == nil {
			return fmt.Errorf("preset resource profile requires a resource pin")
		}
		if err := validatePin(
			document,
			*placement.ResourceProfile,
			resource.KindResourceProfile,
		); err != nil {
			return err
		}
		if placement.ResourceProfile.DomainID != placement.Spec.ResourceProfile.ID {
			return fmt.Errorf("worker dependency resource profile id does not match placement")
		}
	}
	return workerspec.ValidateRuntimePlacement(
		workerspec.RuntimeImage{ID: image.ID, Digest: image.Digest},
		placement.Spec,
	)
}
