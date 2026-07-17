package workerspec

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

type WorkerType struct {
	Slug           slugkit.Slug `json:"slug"`
	DefinitionHash string       `json:"definition_hash"`
}

type RuntimeImage struct {
	ID     int64  `json:"id"`
	Digest string `json:"digest"`
}

type ModelBinding struct {
	ResourceID         int64        `json:"resource_id,omitempty"`
	ResourceRevision   int64        `json:"resource_revision,omitempty"`
	ConnectionID       int64        `json:"connection_id,omitempty"`
	ConnectionRevision int64        `json:"connection_revision,omitempty"`
	ProviderKey        slugkit.Slug `json:"provider_key,omitempty"`
	ProtocolAdapter    slugkit.Slug `json:"protocol_adapter,omitempty"`
	ModelID            string       `json:"model_id,omitempty"`
}

type Runtime struct {
	ModelBinding      ModelBinding       `json:"model_binding"`
	ToolModelBindings []ToolModelBinding `json:"tool_model_bindings,omitempty"`
	WorkerType        WorkerType         `json:"worker_type"`
	Image             RuntimeImage       `json:"image"`
}
