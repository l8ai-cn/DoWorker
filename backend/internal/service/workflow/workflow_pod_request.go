package workflow

import (
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func buildWorkflowCreatePodRequest(
	workflow *workflowDomain.Workflow,
	userID int64,
	agentfileLayer string,
	sourcePodKey string,
	resumeSession bool,
) *agentpodSvc.OrchestrateCreatePodRequest {
	var runnerID int64
	if workflow.RunnerID != nil {
		runnerID = *workflow.RunnerID
	}
	return &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:     workflow.OrganizationID,
		UserID:             userID,
		RunnerID:           runnerID,
		AgentSlug:          workflow.AgentSlug,
		TicketID:           workflow.TicketID,
		ModelResourceID:    workflow.ModelResourceID,
		AgentfileLayer:     &agentfileLayer,
		Cols:               120,
		Rows:               40,
		SourcePodKey:       sourcePodKey,
		ResumeAgentSession: &resumeSession,
	}
}
