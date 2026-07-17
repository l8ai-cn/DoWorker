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
	switch providerKey.String() {
	case "doubao":
		if !strings.HasPrefix(strings.TrimSpace(modelID), "doubao-seedance-") {
			return fmt.Errorf("provider model ID is not a Seedance video model")
		}
	case "sub2api-seedance":
		if strings.TrimSpace(modelID) != "creative-video" {
			return fmt.Errorf("provider model ID is not creative-video")
		}
	}
	return nil
}
