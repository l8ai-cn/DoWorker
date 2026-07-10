package workerruntime

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func ValidateRequest(request Request) error {
	if request.OrganizationID <= 0 {
		return invalidRequest("organization id must be positive")
	}
	if err := slugkit.Validate(request.WorkerTypeSlug.String()); err != nil {
		return fmt.Errorf("%w: worker type slug: %v", ErrInvalidRequest, err)
	}
	if request.RuntimeImageID <= 0 {
		return invalidRequest("runtime image id must be positive")
	}
	if request.ComputeTargetID <= 0 {
		return invalidRequest("compute target id must be positive")
	}
	if request.ResourceProfileID <= 0 {
		return invalidRequest("resource profile id must be positive")
	}
	switch request.PlacementPolicy {
	case workerspec.PlacementPolicyExplicit, workerspec.PlacementPolicyAutomatic:
	default:
		return invalidRequest("placement policy is unsupported")
	}
	switch request.DeploymentMode {
	case workerspec.DeploymentModePooled, workerspec.DeploymentModeDedicated:
	default:
		return invalidRequest("deployment mode is unsupported")
	}
	return nil
}

func invalidRequest(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, reason)
}
