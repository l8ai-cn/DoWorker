package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/serialize"
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

// buildWorkflowAgentfileLayer generates an AgentFile Layer from Workflow configuration.
func (o *WorkflowOrchestrator) buildWorkflowAgentfileLayer(ctx context.Context, workflow *workflowDomain.Workflow, resolvedPrompt string) string {
	var lines []string

	// PROMPT content
	if resolvedPrompt != "" {
		lines = append(lines, fmt.Sprintf("PROMPT %s", serialize.QuoteString(resolvedPrompt)))
	}

	// USE_ENV_BUNDLE bindings — one line per name in the Workflow's ordered
	// list. Backend's ConfigBuilder loads each matching EnvBundle from the
	// user's scope and merges KV into the Pod env in declaration order
	// (later bundles override earlier ones on conflicting keys, mirroring
	// Pod creation). Unknown names are warn-only at eval time, so a
	// renamed/deleted bundle won't fail Workflow creation.
	for _, bundleName := range workflow.UsedEnvBundles {
		if bundleName == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf(`USE_ENV_BUNDLE %s`, serialize.QuoteString(bundleName)))
	}

	// Permission mode
	permissionMode := workflow.PermissionMode
	if permissionMode == "" {
		permissionMode = "bypassPermissions"
	}
	lines = append(lines, fmt.Sprintf(`CONFIG %s = "%s"`, agentDomain.ConfigKeyPermissionMode, permissionMode))

	// Config overrides
	var configOverrides map[string]interface{}
	if workflow.ConfigOverrides != nil {
		_ = json.Unmarshal(workflow.ConfigOverrides, &configOverrides)
	}
	for k, v := range configOverrides {
		if k == agentDomain.ConfigKeyPermissionMode {
			continue // already handled above
		}
		lines = append(lines, fmt.Sprintf("CONFIG %s = %s", k, serialize.FormatValue(v)))
	}

	// Repository slug (resolve from ID)
	if workflow.RepositoryID != nil && o.repoQuery != nil {
		repo, err := o.repoQuery.GetByID(ctx, *workflow.RepositoryID)
		if err == nil && repo != nil {
			lines = append(lines, fmt.Sprintf(`REPO "%s"`, repo.Slug))
			if workflow.BranchName != nil && *workflow.BranchName != "" {
				lines = append(lines, fmt.Sprintf(`BRANCH "%s"`, *workflow.BranchName))
			} else if repo.DefaultBranch != "" {
				lines = append(lines, fmt.Sprintf(`BRANCH "%s"`, repo.DefaultBranch))
			}
		}
	}

	return strings.Join(lines, "\n")
}

// startAutopilot delegates Autopilot creation to AutopilotControllerService.CreateAndStart.
func (o *WorkflowOrchestrator) startAutopilot(ctx context.Context, workflow *workflowDomain.Workflow, run *workflowDomain.WorkflowRun, pod *agentpod.Pod, resolvedPrompt string) (string, error) {
	apCfg := workflow.ParseAutopilotConfig()

	controller, err := o.autopilotSvc.CreateAndStart(ctx, &agentpodSvc.CreateAndStartRequest{
		OrganizationID:      workflow.OrganizationID,
		Pod:                 pod,
		Prompt:              resolvedPrompt,
		MaxIterations:       apCfg.MaxIterations,
		IterationTimeoutSec: apCfg.IterationTimeoutSec,
		NoProgressThreshold: apCfg.NoProgressThreshold,
		SameErrorThreshold:  apCfg.SameErrorThreshold,
		ApprovalTimeoutMin:  apCfg.ApprovalTimeoutMin,
		KeyPrefix:           fmt.Sprintf("workflow-%s-run%d", workflow.Slug, run.RunNumber),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create autopilot controller: %w", err)
	}

	return controller.AutopilotControllerKey, nil
}
