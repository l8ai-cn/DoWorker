package v1

import (
	"github.com/anthropics/agentsmesh/agentfile"
)

const (
	quickTaskPromptMaxLen = 10000
	quickTaskAliasMaxLen  = 100
)

type QuickTaskRequest struct {
	Prompt          string `json:"prompt"`
	AgentSlug       string `json:"agent_slug"`
	RunnerID        int64  `json:"runner_id"`
	RepositoryID    *int64 `json:"repository_id"`
	Alias           string `json:"alias"`
	QueueTTLMinutes int    `json:"queue_ttl_minutes"`
}

type QuickTaskResponse struct {
	PodKey        string `json:"pod_key"`
	Status        string `json:"status"`
	QueuePosition int    `json:"queue_position,omitempty"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

func buildQuickTaskAgentfileLayer(prompt string) string {
	return "PROMPT " + agentfile.FormatStringLiteral(prompt)
}
