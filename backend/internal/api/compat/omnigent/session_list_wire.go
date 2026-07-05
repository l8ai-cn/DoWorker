package omnigent

import (
	"context"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
)

type conversationListItem struct {
	ID                       string            `json:"id"`
	Object                   string            `json:"object"`
	Title                    *string           `json:"title"`
	CreatedAt                int64             `json:"created_at"`
	UpdatedAt                int64             `json:"updated_at"`
	Labels                   map[string]string `json:"labels"`
	PermissionLevel          *int              `json:"permission_level"`
	Owner                    *string           `json:"owner,omitempty"`
	RunnerID                 *string           `json:"runner_id,omitempty"`
	HostID                   *string           `json:"host_id,omitempty"`
	Workspace                *string           `json:"workspace,omitempty"`
	AgentID                  string            `json:"agent_id"`
	AgentName                *string           `json:"agent_name,omitempty"`
	PendingElicitationsCount int               `json:"pending_elicitations_count"`
	Status                   string            `json:"status"`
	RunnerOnline             *bool             `json:"runner_online,omitempty"`
	HostOnline               *bool             `json:"host_online,omitempty"`
	Archived                 bool              `json:"archived"`
	ViewerLastSeen           *int64            `json:"viewer_last_seen,omitempty"`
	ViewerUnread             bool              `json:"viewer_unread,omitempty"`
}

func (d *Deps) listItemFrom(row *domain.Session, pod *podDomain.Pod, online map[string]bool) conversationListItem {
	updated := row.UpdatedAt
	if updated.IsZero() {
		updated = row.CreatedAt
	}
	status := mapSessionStatus(pod)
	item := conversationListItem{
		ID: row.ID, Object: "conversation", Title: row.Title,
		CreatedAt: row.CreatedAt.Unix(), UpdatedAt: updated.Unix(),
		Labels: map[string]string{}, PermissionLevel: nil,
		AgentID: row.AgentSlug, AgentName: strPtr(row.AgentSlug),
		Status: status, Archived: row.Archived,
	}
	item.Labels = sessionLabels(row.Project)
	if d.Elicitations != nil {
		item.PendingElicitationsCount = len(d.Elicitations.PendingPayloads(row.ID))
	}
	if row.RunnerNodeID != nil {
		item.RunnerID = row.RunnerNodeID
		hostID := "host_" + *row.RunnerNodeID
		item.HostID = &hostID
		onlineVal := online[*row.RunnerNodeID]
		item.RunnerOnline = &onlineVal
		hostOnline := onlineVal
		item.HostOnline = &hostOnline
	} else if pod != nil && d.Runner != nil && pod.RunnerID != 0 {
		if r, err := d.Runner.GetRunner(context.Background(), pod.RunnerID); err == nil && r != nil {
			node := r.NodeID
			item.RunnerID = &node
			hostID := "host_" + node
			item.HostID = &hostID
			onlineVal := online[node]
			item.RunnerOnline = &onlineVal
			hostOnline := onlineVal
			item.HostOnline = &hostOnline
		}
	}
	return item
}

func listPageFrom(items []conversationListItem) map[string]any {
	var first, last *string
	if len(items) > 0 {
		first = &items[0].ID
		last = &items[len(items)-1].ID
	}
	return map[string]any{
		"data": items, "first_id": first, "last_id": last, "has_more": false,
	}
}

func mergeSessionGet(item conversationListItem, wire sessionWire) map[string]any {
	return map[string]any{
		"id": item.ID, "object": item.Object, "title": item.Title,
		"created_at": item.CreatedAt, "updated_at": item.UpdatedAt,
		"labels": item.Labels, "permission_level": item.PermissionLevel,
		"agent_id": item.AgentID, "agent_name": item.AgentName,
		"status": item.Status, "runner_id": item.RunnerID,
		"host_id": item.HostID, "runner_online": item.RunnerOnline,
		"host_online": item.HostOnline, "archived": item.Archived,
		"viewer_last_seen": item.ViewerLastSeen, "viewer_unread": item.ViewerUnread,
		"owner": item.Owner,
		"pending_elicitations_count": item.PendingElicitationsCount,
		"harness": wire.Harness, "items": wire.Items,
		"pending_elicitations": wire.PendingElicitations,
		"total_cost_usd": wire.TotalCostUSD, "usage_by_model": wire.UsageByModel,
	}
}