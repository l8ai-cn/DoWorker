package workerspec

import (
	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type ModelRequirement struct {
	Required         bool           `json:"required"`
	ProtocolAdapters []slugkit.Slug `json:"protocol_adapters"`
}

type ToolModelRequirement struct {
	Role             slugkit.Slug
	ProviderKeys     []slugkit.Slug
	ProtocolAdapters []slugkit.Slug
	Modality         resourcedomain.Modality
	Capability       resourcedomain.Capability
	Environment      ToolModelEnvironment
}

func (binding ModelBinding) IsEmpty() bool {
	return binding.ResourceID == 0 &&
		binding.ResourceRevision == 0 &&
		binding.ConnectionID == 0 &&
		binding.ConnectionRevision == 0 &&
		binding.ProviderKey == "" &&
		binding.ProtocolAdapter == "" &&
		binding.ModelID == ""
}
