package agentpod

import (
	"context"
	"log/slog"
)

func (o *PodOrchestrator) resolveTicketID(ctx context.Context, req *OrchestrateCreatePodRequest) {
	if req.TicketID != nil || req.TicketSlug == nil || *req.TicketSlug == "" || o.ticketService == nil {
		return
	}
	t, err := o.ticketService.GetTicketBySlug(ctx, req.OrganizationID, *req.TicketSlug)
	if err == nil && t != nil {
		req.TicketID = &t.ID
	} else if err != nil {
		slog.WarnContext(ctx, "ticket slug resolution failed", "org_id", req.OrganizationID, "ticket_slug", *req.TicketSlug, "error", err)
	}
}
