package sessionapi

import "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"

func sessionCreatePodRequest(
	userID, orgID int64,
	body createSessionBody,
	layer *string,
	workspace string,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:  orgID,
		UserID:          userID,
		AgentSlug:       body.AgentID,
		AgentfileLayer:  layer,
		LocalPath:       workspace,
		ModelResourceID: body.ModelResourceID,
		TokenBudget:     body.TokenBudget,
	}
}
