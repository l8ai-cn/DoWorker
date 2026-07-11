package workercreation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type draftFillPatch struct {
	TypeConfigValues  map[string]any                `json:"type_config_values"`
	InteractionMode   *specdomain.InteractionMode   `json:"interaction_mode"`
	AutomationLevel   *specdomain.AutomationLevel   `json:"automation_level"`
	Branch            *string                       `json:"branch"`
	Instructions      *string                       `json:"instructions"`
	InitialTask       *string                       `json:"initial_task"`
	TerminationPolicy *specdomain.TerminationPolicy `json:"termination_policy"`
	IdleTimeout       *uint32                       `json:"idle_timeout_minutes"`
	Alias             *string                       `json:"alias"`
}

func decodeDraftFillPatch(raw []byte) (draftFillPatch, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	var patch draftFillPatch
	if err := decoder.Decode(&patch); err != nil {
		return draftFillPatch{}, invalidFillField("fill_patch", err.Error())
	}
	var trailing any
	switch err := decoder.Decode(&trailing); {
	case errors.Is(err, io.EOF):
		return patch, nil
	case err == nil:
		return draftFillPatch{}, invalidFillField(
			"fill_patch",
			"contains trailing JSON data",
		)
	default:
		return draftFillPatch{}, invalidFillField("fill_patch", err.Error())
	}
}

func applyDraftFillPatch(current Draft, patch draftFillPatch) (Draft, error) {
	draft, err := cloneWorkerCreationDraft(current)
	if err != nil {
		return Draft{}, err
	}
	if patch.TypeConfigValues != nil {
		if draft.WorkerSpec.TypeConfig.Values == nil {
			draft.WorkerSpec.TypeConfig.Values = map[string]any{}
		}
		for field, value := range patch.TypeConfigValues {
			draft.WorkerSpec.TypeConfig.Values[field] = value
		}
	}
	if patch.InteractionMode != nil {
		draft.WorkerSpec.TypeConfig.InteractionMode = *patch.InteractionMode
	}
	if patch.AutomationLevel != nil {
		draft.WorkerSpec.TypeConfig.AutomationLevel = *patch.AutomationLevel
	}
	if patch.Branch != nil {
		draft.WorkerSpec.Workspace.Branch = *patch.Branch
	}
	if patch.Instructions != nil {
		draft.WorkerSpec.Workspace.Instructions = *patch.Instructions
	}
	if patch.InitialTask != nil {
		draft.WorkerSpec.Workspace.InitialTask = *patch.InitialTask
	}
	if patch.TerminationPolicy != nil {
		draft.WorkerSpec.Lifecycle.TerminationPolicy = *patch.TerminationPolicy
	}
	if patch.IdleTimeout != nil {
		draft.WorkerSpec.Lifecycle.IdleTimeoutMinutes = *patch.IdleTimeout
	}
	if patch.Alias != nil {
		draft.WorkerSpec.Metadata.Alias = *patch.Alias
	}
	return draft, nil
}

func cloneWorkerCreationDraft(current Draft) (Draft, error) {
	cloned := current
	values, err := cloneDraftFillValues(current.WorkerSpec.TypeConfig.Values)
	if err != nil {
		return Draft{}, fmt.Errorf("clone worker draft type config: %w", err)
	}
	cloned.WorkerSpec.TypeConfig.Values = values
	cloned.WorkerSpec.TypeConfig.SecretRefs = make(
		map[string]specdomain.SecretReference,
		len(current.WorkerSpec.TypeConfig.SecretRefs),
	)
	for field, reference := range current.WorkerSpec.TypeConfig.SecretRefs {
		cloned.WorkerSpec.TypeConfig.SecretRefs[field] = reference
	}
	if current.WorkerSpec.Workspace.RepositoryID != nil {
		repositoryID := *current.WorkerSpec.Workspace.RepositoryID
		cloned.WorkerSpec.Workspace.RepositoryID = &repositoryID
	}
	cloned.WorkerSpec.Workspace.SkillIDs = append(
		[]int64{},
		current.WorkerSpec.Workspace.SkillIDs...,
	)
	cloned.WorkerSpec.Workspace.KnowledgeMounts = append(
		[]specdomain.KnowledgeMount{},
		current.WorkerSpec.Workspace.KnowledgeMounts...,
	)
	cloned.WorkerSpec.Workspace.EnvBundleIDs = append(
		[]specdomain.RuntimeEnvBundleID{},
		current.WorkerSpec.Workspace.EnvBundleIDs...,
	)
	if current.WorkerSpec.Metadata.SourceExpertID != nil {
		sourceExpertID := *current.WorkerSpec.Metadata.SourceExpertID
		cloned.WorkerSpec.Metadata.SourceExpertID = &sourceExpertID
	}
	return cloned, nil
}

func cloneDraftFillValues(values map[string]any) (map[string]any, error) {
	if values == nil {
		return nil, nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var cloned map[string]any
	if err := decoder.Decode(&cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}
