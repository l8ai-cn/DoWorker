package airesource

import (
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func ValidateProviderModelCapability(
	providerKey slugkit.Slug,
	modelID string,
	capability Capability,
) error {
	if capability != CapabilityVideoGeneration {
		return nil
	}
	modelID = strings.TrimSpace(modelID)
	switch providerKey.String() {
	case "doubao":
		if strings.HasPrefix(modelID, "doubao-seedance-") {
			return nil
		}
	case "sub2api-seedance":
		if modelID == "doubao-seedance-2-0-260128" {
			return nil
		}
	default:
		return nil
	}
	return fmt.Errorf("provider model ID is not a Seedance video model")
}
