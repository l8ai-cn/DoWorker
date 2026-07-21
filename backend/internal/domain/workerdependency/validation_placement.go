package workerdependency

import (
	"fmt"
	"strings"

	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
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
	if err := validateRuntimeImageReference(image.Reference, image.Digest); err != nil {
		return err
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

func validateRuntimeImageReference(value, digest string) error {
	named, err := reference.ParseNormalizedNamed(value)
	if err == nil {
		canonical, ok := named.(reference.Canonical)
		if ok && canonical.Digest().String() == digest {
			return nil
		}
		return fmt.Errorf("worker dependency runtime image reference must match its digest")
	}
	if localDigest, ok := dockerDaemonReferenceDigest(value); ok {
		if localDigest == digest {
			return nil
		}
		return fmt.Errorf("worker dependency runtime image reference must match its digest")
	}
	return fmt.Errorf("worker dependency runtime image reference is invalid: %w", err)
}

func dockerDaemonReferenceDigest(value string) (string, bool) {
	image, digest, ok := strings.Cut(strings.TrimPrefix(value, "docker-daemon://"), "@")
	if !strings.HasPrefix(value, "docker-daemon://") || !ok || image == "" {
		return "", false
	}
	return digest, digestPattern.MatchString(digest)
}
