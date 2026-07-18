package podconnect

import specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"

type workerTypeSchemaJSON struct {
	Version                     uint32                                 `json:"version"`
	Fields                      map[string]workerTypeFieldSchemaJSON   `json:"fields"`
	CredentialRequirementGroups []workerCredentialRequirementGroupJSON `json:"credential_requirement_groups,omitempty"`
}

type workerTypeFieldSchemaJSON struct {
	Kind        specdomain.TypeFieldKind `json:"kind"`
	Options     []string                 `json:"options,omitempty"`
	Default     any                      `json:"default,omitempty"`
	Required    bool                     `json:"required,omitempty"`
	Description string                   `json:"description,omitempty"`
}

type workerCredentialRequirementGroupJSON struct {
	ID    string   `json:"id"`
	AnyOf []string `json:"any_of"`
}
