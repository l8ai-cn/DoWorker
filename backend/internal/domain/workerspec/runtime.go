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

type Runtime struct {
	ModelResourceID int64        `json:"model_resource_id"`
	WorkerType      WorkerType   `json:"worker_type"`
	Image           RuntimeImage `json:"image"`
}
