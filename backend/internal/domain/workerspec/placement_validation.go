package workerspec

import (
	"fmt"
	"regexp"
	"strings"
)

var runtimeImageDigestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func NormalizeRuntimePlacement(
	image RuntimeImage,
	placement Placement,
) (RuntimeImage, Placement) {
	image.Digest = strings.TrimSpace(image.Digest)
	return image, clonePlacement(placement)
}

func NormalizeAndValidateRuntimePlacement(
	image RuntimeImage,
	placement Placement,
) (RuntimeImage, Placement, error) {
	image, placement = NormalizeRuntimePlacement(image, placement)
	if err := ValidateRuntimePlacement(image, placement); err != nil {
		return RuntimeImage{}, Placement{}, err
	}
	return image, placement, nil
}

func ValidateRuntimePlacement(image RuntimeImage, placement Placement) error {
	if err := validateRuntimeImage(image); err != nil {
		return err
	}
	return validatePlacement(placement)
}

func validateRuntimeImage(image RuntimeImage) error {
	if image.ID <= 0 {
		return fmt.Errorf("runtime image id must be positive")
	}
	if !runtimeImageDigestPattern.MatchString(image.Digest) {
		return fmt.Errorf("runtime image digest must be an immutable SHA-256 digest")
	}
	return nil
}

func validatePlacement(placement Placement) error {
	switch placement.Policy {
	case PlacementPolicyExplicit, PlacementPolicyAutomatic:
	default:
		return fmt.Errorf("invalid placement policy %q", placement.Policy)
	}
	if placement.ComputeTarget.ID <= 0 {
		return fmt.Errorf("compute target id must be positive")
	}
	switch placement.ComputeTarget.Kind {
	case ComputeTargetKindRunnerPool, ComputeTargetKindKubernetes:
	default:
		return fmt.Errorf("unsupported compute target kind %q", placement.ComputeTarget.Kind)
	}
	switch placement.DeploymentMode {
	case DeploymentModePooled, DeploymentModeDedicated:
	default:
		return fmt.Errorf("invalid deployment mode %q", placement.DeploymentMode)
	}
	return validateResourceProfile(placement.ResourceProfile)
}

func validateResourceProfile(profile ResourceProfile) error {
	if profile.ID <= 0 {
		return fmt.Errorf("resource profile id must be positive")
	}
	resources := profile.Resources
	if resources.CPURequestMilliCPU == 0 {
		return fmt.Errorf("cpu request must be positive")
	}
	if resources.CPULimitMilliCPU == 0 {
		return fmt.Errorf("cpu limit must be positive")
	}
	if resources.CPURequestMilliCPU > resources.CPULimitMilliCPU {
		return fmt.Errorf("cpu request must not exceed cpu limit")
	}
	if resources.MemoryRequestBytes == 0 {
		return fmt.Errorf("memory request must be positive")
	}
	if resources.MemoryLimitBytes == 0 {
		return fmt.Errorf("memory limit must be positive")
	}
	if resources.MemoryRequestBytes > resources.MemoryLimitBytes {
		return fmt.Errorf("memory request must not exceed memory limit")
	}
	if (resources.GPURequest == nil) != (resources.GPULimit == nil) {
		return fmt.Errorf("gpu request and limit must be set together")
	}
	if resources.GPURequest == nil {
		return nil
	}
	if *resources.GPURequest == 0 {
		return fmt.Errorf("gpu request must be positive")
	}
	if *resources.GPULimit == 0 {
		return fmt.Errorf("gpu limit must be positive")
	}
	if *resources.GPURequest > *resources.GPULimit {
		return fmt.Errorf("gpu request must not exceed gpu limit")
	}
	return nil
}
