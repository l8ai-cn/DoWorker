package tools

import (
	"encoding/json"
	"fmt"
)

type WorkflowCreateRequest struct {
	Resource json.RawMessage `json:"resource"`
	Enabled  bool            `json:"enabled,omitempty"`
}

type WorkflowCreateResult struct {
	Workflow *WorkflowSummary        `json:"workflow"`
	Resource *AppliedResourceSummary `json:"resource,omitempty"`
}

func (r *WorkflowCreateResult) FormatText() string {
	if r.Workflow == nil {
		return "Workflow created."
	}
	l := r.Workflow
	text := fmt.Sprintf("Workflow created: %s (slug: %s)\nStatus: %s | Mode: %s",
		l.Name, l.Slug, l.Status, l.ExecutionMode)
	if l.CronExpression != "" {
		text += fmt.Sprintf("\nSchedule: %s", l.CronExpression)
		if l.NextRunAt != "" {
			text += fmt.Sprintf(" (next run: %s)", l.NextRunAt)
		}
	} else {
		text += "\nSchedule: on-demand (manual or API trigger)"
	}
	if l.Status == "disabled" {
		text += "\nThe workflow was created disabled. The user can enable it from the Loops page, or call create_workflow with enabled=true after explicit user confirmation."
	}
	if resource := r.Resource.FormatText(); resource != "" {
		text += "\n" + resource
	}
	return text
}
