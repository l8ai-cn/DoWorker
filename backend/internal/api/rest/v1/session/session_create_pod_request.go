package sessionapi

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func (d *Deps) sessionCreatePodRequest(
	ctx context.Context,
	userID, orgID int64,
	orgSlug string,
	body createSessionBody,
	layer *string,
	workspace string,
) (*agentpod.OrchestrateCreatePodRequest, error) {
	draft, err := d.buildFreshWorkerPlan(
		ctx,
		orgID,
		userID,
		orgSlug,
		sessionWorkerPlanInput{
			WorkerSpec:      body.WorkerSpec,
			WorkerTypeSlug:  body.AgentID,
			ModelResourceID: body.ModelResourceID,
			AgentfileLayer:  layer,
			AutomationLevel: body.AutomationLevel,
		},
	)
	if err != nil {
		return nil, err
	}
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:  orgID,
		UserID:          userID,
		LocalPath:       workspace,
		TokenBudget:     body.TokenBudget,
		WorkerSpecDraft: draft,
	}, nil
}
