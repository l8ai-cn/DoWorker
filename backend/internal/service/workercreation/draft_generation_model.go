package workercreation

import (
	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

func draftGenerationModelRequirements() resourceservice.ResolutionRequirements {
	return resourceservice.ResolutionRequirements{
		Modality:   resourcedomain.ModalityChat,
		Capability: resourcedomain.CapabilityTextGeneration,
		AllowedProtocolAdapters: []string{
			"openai-compatible",
			"anthropic",
			"gemini",
		},
	}
}
