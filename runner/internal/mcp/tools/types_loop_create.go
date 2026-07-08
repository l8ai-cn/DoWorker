package tools

import "fmt"

// LoopCreateRequest carries the fields an agent may set when persisting a
// clarified loop design via the create_loop MCP method. Agent/runner default
// to the calling pod's when omitted; the backend owns validation.
type LoopCreateRequest struct {
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

// LoopCreateResult wraps the created loop summary returned by the backend.
type LoopCreateResult struct {
	Loop *LoopSummary `json:"loop"`
}

func (r *LoopCreateResult) FormatText() string {
	if r.Loop == nil {
		return "Loop created."
	}
	l := r.Loop
	text := fmt.Sprintf("Loop created: %s (slug: %s)\nStatus: %s | Mode: %s",
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
		text += "\nThe loop was created disabled. The user can enable it from the Loops page, or call create_loop with enabled=true after explicit user confirmation."
	}
	return text
}
