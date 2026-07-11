package workflow

import (
	"context"
	"log/slog"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
)

func (s *WorkflowRunService) resolveRunStatuses(ctx context.Context, runs []*workflowDomain.WorkflowRun) {
	podKeys := make([]string, 0)
	autopilotKeys := make([]string, 0)
	for _, run := range runs {
		if run.PodKey != nil {
			podKeys = append(podKeys, *run.PodKey)
		}
		if run.AutopilotControllerKey != nil {
			autopilotKeys = append(autopilotKeys, *run.AutopilotControllerKey)
		}
	}

	if len(podKeys) == 0 {
		return
	}

	podInfos, err := s.repo.BatchGetPodStatuses(ctx, podKeys)
	if err != nil {
		slog.ErrorContext(ctx, "failed to batch get pod statuses for SSOT resolution", "error", err, "count", len(podKeys))
	}
	podMap := make(map[string]*workflowDomain.PodStatusInfo, len(podInfos))
	for i := range podInfos {
		podMap[podInfos[i].PodKey] = &podInfos[i]
	}

	autopilotMap, err := s.repo.BatchGetAutopilotPhases(ctx, autopilotKeys)
	if err != nil {
		slog.ErrorContext(ctx, "failed to batch get autopilot phases for SSOT resolution", "error", err, "count", len(autopilotKeys))
	}

	for _, run := range runs {
		resolveOneRunStatus(run, podMap, autopilotMap)
	}
}

func resolveOneRunStatus(run *workflowDomain.WorkflowRun, podMap map[string]*workflowDomain.PodStatusInfo, autopilotMap map[string]string) {
	if run.PodKey == nil {
		return
	}
	if run.FinishedAt != nil {
		return
	}
	pod, ok := podMap[*run.PodKey]
	if !ok {
		run.Status = workflowDomain.RunStatusFailed
		return
	}

	autopilotPhase := ""
	if run.AutopilotControllerKey != nil && autopilotMap != nil {
		autopilotPhase = autopilotMap[*run.AutopilotControllerKey]
	}

	ResolveRunStatus(run, pod.Status, autopilotPhase, pod.FinishedAt)
}

func (s *WorkflowRunService) resolveRunStatus(ctx context.Context, run *workflowDomain.WorkflowRun) {
	if run.PodKey == nil {
		return
	}
	s.resolveRunStatuses(ctx, []*workflowDomain.WorkflowRun{run})
}
