package v1

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func hasRESTResumeRuntimeOverrides(req CreatePodRequest) bool {
	return req.AgentSlug != "" ||
		req.RunnerID != 0 ||
		req.RepositoryID != nil ||
		req.Alias != nil ||
		req.AgentfileLayer != nil ||
		req.AutomationLevel != "" ||
		req.ModelResourceID != nil ||
		req.TokenBudget != nil ||
		req.Perpetual != nil ||
		len(req.KnowledgeMounts) > 0
}

func buildRESTResumePodRequest(
	req CreatePodRequest,
	tenant *middleware.TenantContext,
) *agentpod.OrchestrateCreatePodRequest {
	result := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:     tenant.OrganizationID,
		UserID:             tenant.UserID,
		TicketSlug:         req.TicketSlug,
		Cols:               req.Cols,
		Rows:               req.Rows,
		SourcePodKey:       req.SourcePodKey,
		ResumeAgentSession: req.ResumeAgentSession,
		QueueIfUnavailable: req.QueueIfOffline,
	}
	if req.QueueIfOffline {
		ttl := req.QueueTTLMinutes
		if ttl == 0 {
			ttl = 30
		}
		result.QueueTTL = time.Duration(ttl) * time.Minute
	}
	return result
}
