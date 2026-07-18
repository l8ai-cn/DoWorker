package workflow

import (
	"encoding/json"
	"errors"
)

var (
	ErrWorkflowRunExecutionManifestRequired = errors.New(
		"workflow run execution manifest is required",
	)
	ErrWorkflowRunExecutionManifestCorrupt = errors.New(
		"workflow run execution manifest is corrupt",
	)
)

type WorkflowRunExecutionManifest struct {
	Version            int                   `json:"version"`
	OrganizationID     int64                 `json:"organization_id"`
	WorkflowName       string                `json:"workflow_name"`
	WorkflowSlug       string                `json:"workflow_slug"`
	CreatedByID        int64                 `json:"created_by_id"`
	ExecutionMode      string                `json:"execution_mode"`
	Autopilot          AutopilotConfigValues `json:"autopilot"`
	SandboxStrategy    string                `json:"sandbox_strategy"`
	SessionPersistence bool                  `json:"session_persistence"`
	SourcePodKey       string                `json:"source_pod_key,omitempty"`
	CallbackURL        string                `json:"callback_url,omitempty"`
	TicketID           *int64                `json:"ticket_id,omitempty"`
	MaxRetainedRuns    int                   `json:"max_retained_runs"`
	TimeoutMinutes     int                   `json:"timeout_minutes"`
	IdleTimeoutSeconds int                   `json:"idle_timeout_seconds"`
}

func PinWorkflowRunExecutionManifest(item *Workflow) (json.RawMessage, error) {
	if item == nil {
		return nil, ErrWorkflowRunExecutionManifestCorrupt
	}
	var autopilot AutopilotConfigValues
	if len(item.AutopilotConfig) > 0 {
		if err := json.Unmarshal(item.AutopilotConfig, &autopilot); err != nil {
			return nil, ErrWorkflowRunExecutionManifestCorrupt
		}
	}
	manifest := WorkflowRunExecutionManifest{
		Version:            1,
		OrganizationID:     item.OrganizationID,
		WorkflowName:       item.Name,
		WorkflowSlug:       item.Slug,
		CreatedByID:        item.CreatedByID,
		ExecutionMode:      item.ExecutionMode,
		Autopilot:          autopilot,
		SandboxStrategy:    item.SandboxStrategy,
		SessionPersistence: item.SessionPersistence,
		CallbackURL:        optionalString(item.CallbackURL),
		TicketID:           cloneInt64(item.TicketID),
		MaxRetainedRuns:    item.MaxRetainedRuns,
		TimeoutMinutes:     item.TimeoutMinutes,
		IdleTimeoutSeconds: item.IdleTimeoutSec,
	}
	if item.IsPersistent() && item.LastPodKey != nil {
		manifest.SourcePodKey = *item.LastPodKey
	}
	if !manifest.valid() {
		return nil, ErrWorkflowRunExecutionManifestCorrupt
	}
	content, err := json.Marshal(manifest)
	if err != nil {
		return nil, ErrWorkflowRunExecutionManifestCorrupt
	}
	return content, nil
}

func (run *WorkflowRun) PinnedExecution() (
	WorkflowRunExecutionManifest,
	error,
) {
	if run == nil || len(run.ExecutionManifest) == 0 {
		return WorkflowRunExecutionManifest{},
			ErrWorkflowRunExecutionManifestRequired
	}
	var manifest WorkflowRunExecutionManifest
	if err := json.Unmarshal(run.ExecutionManifest, &manifest); err != nil ||
		!manifest.valid() {
		return WorkflowRunExecutionManifest{},
			ErrWorkflowRunExecutionManifestCorrupt
	}
	return manifest, nil
}

func (manifest WorkflowRunExecutionManifest) valid() bool {
	if manifest.Version != 1 ||
		manifest.OrganizationID <= 0 ||
		manifest.WorkflowName == "" ||
		manifest.WorkflowSlug == "" ||
		manifest.CreatedByID <= 0 ||
		manifest.MaxRetainedRuns < 0 ||
		manifest.TimeoutMinutes <= 0 ||
		manifest.IdleTimeoutSeconds < 0 {
		return false
	}
	if manifest.ExecutionMode != ExecutionModeDirect &&
		manifest.ExecutionMode != ExecutionModeAutopilot {
		return false
	}
	if manifest.SandboxStrategy != SandboxStrategyFresh &&
		manifest.SandboxStrategy != SandboxStrategyPersistent {
		return false
	}
	if manifest.SandboxStrategy == SandboxStrategyFresh &&
		(manifest.SessionPersistence || manifest.SourcePodKey != "") {
		return false
	}
	return manifest.TicketID == nil || *manifest.TicketID > 0
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
