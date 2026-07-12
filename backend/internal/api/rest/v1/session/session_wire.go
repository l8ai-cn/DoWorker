package sessionapi

import (
	"context"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
)

type sessionWire struct {
	ID                       string           `json:"id"`
	AgentID                  string           `json:"agent_id"`
	AgentName                *string          `json:"agent_name,omitempty"`
	RunnerID                 *string          `json:"runner_id,omitempty"`
	Status                   string           `json:"status"`
	CreatedAt                int64            `json:"created_at"`
	Title                    *string          `json:"title,omitempty"`
	Items                    []any            `json:"items,omitempty"`
	Harness                  *string          `json:"harness,omitempty"`
	InteractionMode          string           `json:"interaction_mode,omitempty"`
	PendingElicitations      []map[string]any `json:"pending_elicitations,omitempty"`
	PendingElicitationsCount int              `json:"pending_elicitations_count,omitempty"`
	TotalCostUSD             *float64         `json:"total_cost_usd,omitempty"`
	UsageByModel             map[string]any   `json:"usage_by_model,omitempty"`
}

func sessionWireFrom(row *domain.Session, pod *podDomain.Pod, runnerNodeID *string, pending []map[string]any, agg sessionusagesvc.Aggregate) sessionWire {
	w := sessionWire{
		ID: row.ID, AgentID: row.AgentSlug, AgentName: strPtr(row.AgentSlug),
		Status: mapSessionStatus(pod), CreatedAt: row.CreatedAt.Unix(),
		Title: row.Title, Items: []any{}, Harness: strPtr(row.AgentSlug),
		TotalCostUSD: agg.TotalCostUSD,
	}
	if len(agg.UsageByModel) > 0 {
		byModel := make(map[string]any, len(agg.UsageByModel))
		for k, v := range agg.UsageByModel {
			byModel[k] = v
		}
		w.UsageByModel = byModel
	}
	if runnerNodeID != nil {
		w.RunnerID = runnerNodeID
	} else if row.RunnerNodeID != nil {
		w.RunnerID = row.RunnerNodeID
	}
	if pod != nil {
		w.InteractionMode = pod.InteractionMode
	}
	if len(pending) > 0 {
		w.PendingElicitations = pending
		w.PendingElicitationsCount = len(pending)
	}
	return w
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func (d *Deps) sessionWire(row *domain.Session, pod *podDomain.Pod, runnerNodeID *string) sessionWire {
	var pending []map[string]any
	if d.Elicitations != nil {
		pending = d.Elicitations.PendingPayloads(row.ID)
	}
	var agg sessionusagesvc.Aggregate
	if d.SessionUsage != nil && pod != nil && pod.PodKey != "" {
		agg, _ = d.SessionUsage.Aggregate(context.Background(), pod.PodKey)
	}
	return sessionWireFrom(row, pod, runnerNodeID, pending, agg)
}
