package workflow

import (
	"errors"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

var ErrWorkflowResourceBindingCorrupt = errors.New(
	"workflow orchestration resource binding is corrupt",
)

func buildWorkflowRunPodRequest(
	workflow *workflowDomain.Workflow,
	run *workflowDomain.WorkflowRun,
	userID int64,
	resolvedPrompt string,
	agentfileLayer string,
	sourcePodKey string,
	resumeSession bool,
) (*agentpodSvc.OrchestrateCreatePodRequest, error) {
	if !workflow.IsResourceManaged() {
		return buildWorkflowCreatePodRequest(
			workflow,
			userID,
			agentfileLayer,
			sourcePodKey,
			resumeSession,
		), nil
	}
	if !validWorkflowRunResourceBinding(workflow, run) {
		return nil, ErrWorkflowResourceBindingCorrupt
	}
	snapshotID := *run.WorkerSpecSnapshotID
	prompt := resolvedPrompt
	return &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:           workflow.OrganizationID,
		UserID:                   userID,
		WorkerSpecSnapshotID:     &snapshotID,
		WorkerSpecPromptOverride: &prompt,
		Cols:                     120,
		Rows:                     40,
		SourcePodKey:             sourcePodKey,
		ResumeAgentSession:       &resumeSession,
	}, nil
}

func validWorkflowRunResourceBinding(
	workflow *workflowDomain.Workflow,
	run *workflowDomain.WorkflowRun,
) bool {
	if workflow.OrchestrationResourceID == nil ||
		*workflow.OrchestrationResourceID <= 0 ||
		workflow.OrchestrationResourceRevision == nil ||
		*workflow.OrchestrationResourceRevision <= 0 ||
		workflow.WorkerSpecSnapshotID == nil ||
		*workflow.WorkerSpecSnapshotID <= 0 ||
		run.OrchestrationResourceID == nil ||
		run.OrchestrationResourceRevision == nil ||
		run.WorkerSpecSnapshotID == nil {
		return false
	}
	return *workflow.OrchestrationResourceID ==
		*run.OrchestrationResourceID &&
		*workflow.OrchestrationResourceRevision ==
			*run.OrchestrationResourceRevision &&
		*workflow.WorkerSpecSnapshotID == *run.WorkerSpecSnapshotID
}

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
