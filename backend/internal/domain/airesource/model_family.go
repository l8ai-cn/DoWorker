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
	if isSeedanceVideoProvider(providerKey.String()) &&
		capability == CapabilityVideoGeneration &&
		!strings.HasPrefix(strings.TrimSpace(modelID), "doubao-seedance-") {
		return fmt.Errorf("provider model ID is not a Seedance video model")
	}
	return nil
}

func isSeedanceVideoProvider(providerKey string) bool {
	return providerKey == "doubao" || providerKey == "sub2api-seedance"
}
