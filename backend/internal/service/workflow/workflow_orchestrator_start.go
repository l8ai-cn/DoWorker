package workflow

import (
	"context"
	"fmt"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
)

func (o *WorkflowOrchestrator) StartRun(ctx context.Context, workflow *workflowDomain.Workflow, run *workflowDomain.WorkflowRun, userID int64) {
	defer func() {
		if r := recover(); r != nil {
			o.logger.Error("panic in StartRun", "run_id", run.ID, "workflow_id", workflow.ID, "panic", r)
			_ = o.MarkRunFailed(ctx, run.ID, fmt.Sprintf("Internal error: %v", r))
		}
	}()

	if o.podOrchestrator == nil {
		o.logger.Error("pod orchestrator not set, cannot start run", "run_id", run.ID)
		_ = o.MarkRunFailed(ctx, run.ID, "Pod orchestrator not configured")
		return
	}

	currentRun, err := o.workflowRunService.GetByID(ctx, run.ID)
	if err != nil {
		o.logger.Error("failed to check run status before start", "run_id", run.ID, "error", err)
		return
	}
	if currentRun.FinishedAt != nil || currentRun.IsTerminal() {
		o.logger.Info("run already finished/cancelled before StartRun, skipping",
			"run_id", run.ID, "status", currentRun.Status)
		return
	}

	resolvedPrompt := resolvePrompt(workflow.PromptTemplate, workflow.PromptVariables, run.TriggerParams)
	if err := o.workflowRunService.UpdateStatus(ctx, run.ID, map[string]interface{}{
		"resolved_prompt": resolvedPrompt,
	}); err != nil {
		o.logger.Error("failed to persist resolved prompt", "run_id", run.ID, "error", err)
	}

	agentfileLayer := ""
	if !workflow.IsResourceManaged() {
		agentfileLayer = o.buildWorkflowAgentfileLayer(
			ctx,
			workflow,
			resolvedPrompt,
		)
	}

	var sourcePodKey string
	resumeSession := workflow.SessionPersistence
	if workflow.IsPersistent() && workflow.LastPodKey != nil {
		sourcePodKey = *workflow.LastPodKey
	}

	podRequest, err := buildWorkflowRunPodRequest(
		workflow,
		currentRun,
		userID,
		resolvedPrompt,
		agentfileLayer,
		sourcePodKey,
		resumeSession,
	)
	if err != nil {
		_ = o.MarkRunFailed(ctx, run.ID, err.Error())
		return
	}
	podResult, err := o.podOrchestrator.CreatePod(ctx, podRequest)
	if err != nil {
		_ = o.MarkRunFailed(ctx, run.ID, fmt.Sprintf("Pod creation failed: %v", err))
		return
	}

	pod := podResult.Pod
	autopilotKey := ""

	if workflow.IsAutopilot() && o.autopilotSvc != nil {
		var err error
		autopilotKey, err = o.startAutopilot(ctx, workflow, run, pod, resolvedPrompt)
		if err != nil {
			o.logger.Error("autopilot creation failed, terminating Pod",
				"run_id", run.ID, "pod_key", pod.PodKey, "error", err)
			if o.podTerminator != nil {
				_ = o.podTerminator.TerminatePod(ctx, pod.PodKey)
			}
			_ = o.MarkRunFailed(ctx, run.ID, fmt.Sprintf("Autopilot creation failed: %v", err))
			return
		}
	}

	if err := o.SetRunPodKey(ctx, run.ID, pod.PodKey, autopilotKey); err != nil {
		o.logger.Error("failed to set run pod key", "run_id", run.ID, "error", err)
	}
	// Re-publish workflow_run:started now that pod_key is bound to the run.
	// The earlier publish in TriggerRun fired before pod creation, so
	// subscribers (web + desktop realtime e2e) saw an empty pod_key and
	// couldn't correlate the run to a pod for completion detection.
	// Reload the run to capture the persisted PodKey before publishing.
	if updated, err := o.workflowRunService.GetByID(ctx, run.ID); err == nil {
		o.publishRunEvent(workflow.OrganizationID, eventbus.EventWorkflowRunStarted, updated)
	} else {
		o.logger.Warn("failed to reload run after SetRunPodKey for republish", "run_id", run.ID, "error", err)
	}

	o.logger.Info("workflow run started",
		"workflow_id", workflow.ID,
		"run_id", run.ID,
		"pod_key", pod.PodKey,
		"autopilot_key", autopilotKey,
		"execution_mode", workflow.ExecutionMode,
	)
}
