package workerspec

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

type InteractionMode string

const (
	InteractionModePTY InteractionMode = "pty"
	InteractionModeACP InteractionMode = "acp"
)

type AutomationLevel string

const (
	AutomationLevelInteractive AutomationLevel = "interactive"
	AutomationLevelAutoEdit    AutomationLevel = "auto_edit"
	AutomationLevelAutonomous  AutomationLevel = "autonomous"
)

type SecretReference struct {
	Kind slugkit.Slug `json:"kind"`
	ID   int64        `json:"id"`
}

type TypeConfig struct {
	SchemaVersion   uint32                     `json:"schema_version"`
	Values          map[string]any             `json:"values"`
	SecretRefs      map[string]SecretReference `json:"secret_refs"`
	InteractionMode InteractionMode            `json:"interaction_mode"`
	AutomationLevel AutomationLevel            `json:"automation_level"`
}
