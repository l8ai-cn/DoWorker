package tools

import "fmt"

// WorkflowCreateRequest carries the fields an agent may set when persisting a
// clarified workflow design via the create_workflow MCP method. Agent/runner default
// to the calling pod's when omitted; the backend owns validation.
type WorkflowCreateRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	PromptTemplate     string `json:"prompt_template"`
	AgentSlug          string `json:"agent_slug,omitempty"`
	CronExpression     string `json:"cron_expression,omitempty"`
	ExecutionMode      string `json:"execution_mode,omitempty"`
	SandboxStrategy    string `json:"sandbox_strategy,omitempty"`
	ConcurrencyPolicy  string `json:"concurrency_policy,omitempty"`
	TimeoutMinutes     int    `json:"timeout_minutes,omitempty"`
	MaxConcurrentRuns  int    `json:"max_concurrent_runs,omitempty"`
	MaxRetainedRuns    int    `json:"max_retained_runs,omitempty"`
	SessionPersistence bool   `json:"session_persistence,omitempty"`
	RepositoryID       *int64 `json:"repository_id,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
}

// WorkflowCreateResult wraps the created workflow summary returned by the backend.
type WorkflowCreateResult struct {
	Workflow *WorkflowSummary `json:"workflow"`
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
	return text
}
