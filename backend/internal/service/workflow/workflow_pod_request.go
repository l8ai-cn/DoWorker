package workflow

import (
	"errors"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	agentpodSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
)

var ErrWorkflowResourceBindingCorrupt = errors.New(
	"workflow orchestration resource binding is corrupt",
)

func buildWorkflowRunPodRequest(
	run *workflowDomain.WorkflowRun,
	userID int64,
) (*agentpodSvc.OrchestrateCreatePodRequest, error) {
	manifest, err := validWorkflowRunResourceBinding(run)
	if err != nil {
		return nil, ErrWorkflowResourceBindingCorrupt
	}
	prompt := *run.ResolvedPrompt
	if manifest.SourcePodKey != "" {
		return buildWorkflowRunLineagePodRequest(
			manifest,
			userID,
			prompt,
		), nil
	}
	return buildWorkflowRunSnapshotPodRequest(
		manifest,
		userID,
		*run.WorkerSpecSnapshotID,
		prompt,
	), nil
}

func buildWorkflowRunSnapshotPodRequest(
	manifest workflowDomain.WorkflowRunExecutionManifest,
	userID int64,
	snapshotID int64,
	prompt string,
) *agentpodSvc.OrchestrateCreatePodRequest {
	return &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:           manifest.OrganizationID,
		UserID:                   userID,
		WorkerSpecSnapshotID:     &snapshotID,
		WorkerSpecPromptOverride: &prompt,
		Cols:                     120,
		Rows:                     40,
		ResumeAgentSession:       &manifest.SessionPersistence,
	}
}

func buildWorkflowRunLineagePodRequest(
	manifest workflowDomain.WorkflowRunExecutionManifest,
	userID int64,
	prompt string,
) *agentpodSvc.OrchestrateCreatePodRequest {
	return &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:           manifest.OrganizationID,
		UserID:                   userID,
		WorkerSpecPromptOverride: &prompt,
		Cols:                     120,
		Rows:                     40,
		SourcePodKey:             manifest.SourcePodKey,
		ResumeAgentSession:       &manifest.SessionPersistence,
	}
}

func validWorkflowRunResourceBinding(
	run *workflowDomain.WorkflowRun,
) (workflowDomain.WorkflowRunExecutionManifest, error) {
	if run == nil ||
		run.OrganizationID <= 0 ||
		run.OrchestrationResourceID == nil ||
		*run.OrchestrationResourceID <= 0 ||
		run.OrchestrationResourceRevision == nil ||
		*run.OrchestrationResourceRevision <= 0 ||
		run.WorkerSpecSnapshotID == nil ||
		*run.WorkerSpecSnapshotID <= 0 ||
		run.ResolvedPrompt == nil {
		return workflowDomain.WorkflowRunExecutionManifest{},
			ErrWorkflowResourceBindingCorrupt
	}
	manifest, err := run.PinnedExecution()
	if err != nil || manifest.OrganizationID != run.OrganizationID {
		return workflowDomain.WorkflowRunExecutionManifest{},
			ErrWorkflowResourceBindingCorrupt
	}
	return manifest, nil
}
