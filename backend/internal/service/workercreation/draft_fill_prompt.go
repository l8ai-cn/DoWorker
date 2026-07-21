package workercreation

import (
	"encoding/json"
	"fmt"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

const draftFillSystemPrompt = `Fill an existing Worker draft from the user's request.
Return exactly one JSON object without markdown or explanatory text.
Allowed keys: type_config_values, interaction_mode, automation_level, branch, instructions, initial_task, termination_policy, idle_timeout_minutes, alias.
Never output model, runtime, placement, repository, Skill, knowledge, environment bundle, Expert, credential, secret reference, or identifier fields.
type_config_values is merged into the existing values and may only contain non-secret fields declared by the supplied schema.
interaction_mode must use a supplied supported value.
automation_level must be interactive, auto_edit, or autonomous.
termination_policy must be manual, idle, or completed. idle requires a positive idle_timeout_minutes; manual and completed require zero.
Use JSON booleans and numbers for typed fields. Preserve existing values unless the request clearly changes them.`

type draftFillPromptSchema struct {
	Version uint32                          `json:"version"`
	Fields  map[string]draftFillPromptField `json:"fields"`
}

type draftFillPromptField struct {
	Kind    specdomain.TypeFieldKind `json:"kind"`
	Options []string                 `json:"options,omitempty"`
}

type draftFillPromptCurrent struct {
	TypeConfigValues map[string]any               `json:"type_config_values"`
	InteractionMode  specdomain.InteractionMode   `json:"interaction_mode"`
	AutomationLevel  specdomain.AutomationLevel   `json:"automation_level"`
	RepositorySet    bool                         `json:"repository_selected"`
	Branch           string                       `json:"branch"`
	Instructions     string                       `json:"instructions"`
	InitialTask      string                       `json:"initial_task"`
	Termination      specdomain.TerminationPolicy `json:"termination_policy"`
	IdleTimeout      uint32                       `json:"idle_timeout_minutes"`
	Alias            string                       `json:"alias"`
}

type draftFillPromptRequest struct {
	Request                   string                       `json:"request"`
	WorkerType                string                       `json:"worker_type"`
	Schema                    draftFillPromptSchema        `json:"type_schema"`
	SupportedInteractionModes []specdomain.InteractionMode `json:"supported_interaction_modes"`
	Current                   draftFillPromptCurrent       `json:"current"`
}

func buildDraftFillPrompts(
	request string,
	current Draft,
	workerType specservice.WorkerTypeResolution,
) (string, string, error) {
	fields := make(
		map[string]draftFillPromptField,
		len(workerType.TypeSchema.Fields),
	)
	for name, field := range workerType.TypeSchema.Fields {
		fields[name] = draftFillPromptField{
			Kind:    field.Kind,
			Options: append([]string{}, field.Options...),
		}
	}
	payload := draftFillPromptRequest{
		Request:    request,
		WorkerType: current.WorkerSpec.WorkerTypeSlug.String(),
		Schema: draftFillPromptSchema{
			Version: workerType.TypeSchema.Version,
			Fields:  fields,
		},
		SupportedInteractionModes: append(
			[]specdomain.InteractionMode{},
			workerType.SupportedInteractionModes...,
		),
		Current: draftFillPromptCurrent{
			TypeConfigValues: current.WorkerSpec.TypeConfig.Values,
			InteractionMode:  current.WorkerSpec.TypeConfig.InteractionMode,
			AutomationLevel:  current.WorkerSpec.TypeConfig.AutomationLevel,
			RepositorySet:    current.WorkerSpec.Workspace.RepositoryID != nil,
			Branch:           current.WorkerSpec.Workspace.Branch,
			Instructions:     current.WorkerSpec.Workspace.Instructions,
			InitialTask:      current.WorkerSpec.Workspace.InitialTask,
			Termination:      current.WorkerSpec.Lifecycle.TerminationPolicy,
			IdleTimeout:      current.WorkerSpec.Lifecycle.IdleTimeoutMinutes,
			Alias:            current.WorkerSpec.Metadata.Alias,
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("encode worker draft fill prompt: %w", err)
	}
	return draftFillSystemPrompt, string(encoded), nil
}
