package workflow

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	agentpodSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
)

func (o *WorkflowOrchestrator) startAutopilot(
	ctx context.Context,
	manifest workflowDomain.WorkflowRunExecutionManifest,
	run *workflowDomain.WorkflowRun,
	pod *agentpod.Pod,
	resolvedPrompt string,
) (string, error) {
	config := manifest.Autopilot
	controller, err := o.autopilotSvc.CreateAndStart(
		ctx,
		&agentpodSvc.CreateAndStartRequest{
			OrganizationID:      manifest.OrganizationID,
			Pod:                 pod,
			Prompt:              resolvedPrompt,
			MaxIterations:       config.MaxIterations,
			IterationTimeoutSec: config.IterationTimeoutSec,
			NoProgressThreshold: config.NoProgressThreshold,
			SameErrorThreshold:  config.SameErrorThreshold,
			ApprovalTimeoutMin:  config.ApprovalTimeoutMin,
			KeyPrefix: fmt.Sprintf(
				"workflow-%s-run%d",
				manifest.WorkflowSlug,
				run.RunNumber,
			),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create autopilot controller: %w", err)
	}
	return controller.AutopilotControllerKey, nil
}
