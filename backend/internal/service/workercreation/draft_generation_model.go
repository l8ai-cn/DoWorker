package workercreation

import (
	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	resourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
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
