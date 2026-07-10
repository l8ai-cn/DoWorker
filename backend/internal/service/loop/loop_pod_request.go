package loop

import (
	loopDomain "github.com/anthropics/agentsmesh/backend/internal/domain/loop"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func buildLoopCreatePodRequest(
	loop *loopDomain.Loop,
	userID int64,
	agentfileLayer string,
	sourcePodKey string,
	resumeSession bool,
) *agentpodSvc.OrchestrateCreatePodRequest {
	var runnerID int64
	if loop.RunnerID != nil {
		runnerID = *loop.RunnerID
	}
	return &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:     loop.OrganizationID,
		UserID:             userID,
		RunnerID:           runnerID,
		AgentSlug:          loop.AgentSlug,
		TicketID:           loop.TicketID,
		ModelResourceID:    loop.ModelResourceID,
		AgentfileLayer:     &agentfileLayer,
		Cols:               120,
		Rows:               40,
		SourcePodKey:       sourcePodKey,
		ResumeAgentSession: &resumeSession,
	}
}
