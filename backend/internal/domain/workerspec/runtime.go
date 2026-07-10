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
	ResourceID         int64        `json:"resource_id"`
	ResourceRevision   int64        `json:"resource_revision"`
	ConnectionID       int64        `json:"connection_id"`
	ConnectionRevision int64        `json:"connection_revision"`
	ProviderKey        slugkit.Slug `json:"provider_key"`
	ModelID            string       `json:"model_id"`
}

type Runtime struct {
	ModelBinding ModelBinding `json:"model_binding"`
	WorkerType   WorkerType   `json:"worker_type"`
	Image        RuntimeImage `json:"image"`
}
