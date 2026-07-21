package workflow

import (
	"context"
	"log/slog"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/lib/pq"
)

func (s *WorkflowService) Update(ctx context.Context, orgID int64, slug string, req *UpdateWorkflowRequest) (*workflowDomain.Workflow, error) {
	workflow, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if workflow.IsResourceManaged() {
		return nil, ErrWorkflowManagedByResourceApply
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.AgentSlug != "" {
		updates["agent_slug"] = req.AgentSlug
	}
	if req.PermissionMode != nil {
		updates["permission_mode"] = *req.PermissionMode
	}
	if req.PromptTemplate != nil {
		updates["prompt_template"] = *req.PromptTemplate
	}
	if req.RepositoryID != nil {
		updates["repository_id"] = *req.RepositoryID
	}
	if req.RunnerID != nil {
		updates["runner_id"] = *req.RunnerID
	}
	if req.BranchName != nil {
		updates["branch_name"] = *req.BranchName
	}
	if req.TicketID != nil {
		updates["ticket_id"] = *req.TicketID
	}
	if req.ModelResourceID != nil {
		updates["model_resource_id"] = *req.ModelResourceID
	}
	if req.UsedEnvBundles != nil {
		// Nil pointer = leave unchanged; pointer to []string = replace.
		// Empty slice replaces with no bundles. pq.StringArray serialises
		// `[]string{}` to PostgreSQL `'{}'::text[]`.
		updates["used_env_bundles"] = pq.StringArray(*req.UsedEnvBundles)
	}
	if req.ConfigOverrides != nil {
		updates["config_overrides"] = req.ConfigOverrides
	}
	if req.ExecutionMode != nil {
		updates["execution_mode"] = *req.ExecutionMode
	}
	if req.CronExpression != nil {
		updates["cron_expression"] = *req.CronExpression
		if *req.CronExpression != "" {
			schedule, err := cronParser.Parse(*req.CronExpression)
			if err != nil {
				return nil, ErrInvalidCron
			}
			next := schedule.Next(time.Now())
			updates["next_run_at"] = next
		} else {
			updates["next_run_at"] = nil
		}
	}
	if req.AutopilotConfig != nil {
		updates["autopilot_config"] = req.AutopilotConfig
	}
	if req.PromptVariables != nil {
		updates["prompt_variables"] = req.PromptVariables
	}
	if req.CallbackURL != nil {
		if *req.CallbackURL == "" {
			updates["callback_url"] = nil
		} else {
			if err := validateCallbackURL(*req.CallbackURL); err != nil {
				return nil, err
			}
			updates["callback_url"] = *req.CallbackURL
		}
	}
	if req.SandboxStrategy != nil {
		updates["sandbox_strategy"] = *req.SandboxStrategy
	}
	if req.SessionPersistence != nil {
		updates["session_persistence"] = *req.SessionPersistence
	}
	if req.ConcurrencyPolicy != nil {
		updates["concurrency_policy"] = *req.ConcurrencyPolicy
	}
	if req.MaxConcurrentRuns != nil {
		updates["max_concurrent_runs"] = *req.MaxConcurrentRuns
	}
	if req.MaxRetainedRuns != nil {
		updates["max_retained_runs"] = *req.MaxRetainedRuns
	}
	if req.TimeoutMinutes != nil {
		updates["timeout_minutes"] = *req.TimeoutMinutes
	}
	if req.IdleTimeoutSec != nil {
		updates["idle_timeout_sec"] = *req.IdleTimeoutSec
	}

	if req.RunnerID != nil {
		effectiveRunnerID := *req.RunnerID
		currentRunnerID := int64(0)
		if workflow.RunnerID != nil {
			currentRunnerID = *workflow.RunnerID
		}
		if effectiveRunnerID != currentRunnerID && workflow.IsPersistent() && workflow.LastPodKey != nil {
			updates["last_pod_key"] = nil
			updates["sandbox_path"] = nil
		}
	}

	if req.SandboxStrategy != nil && *req.SandboxStrategy == workflowDomain.SandboxStrategyFresh &&
		workflow.SandboxStrategy == workflowDomain.SandboxStrategyPersistent {
		updates["last_pod_key"] = nil
		updates["sandbox_path"] = nil
	}

	execMode := ""
	if req.ExecutionMode != nil {
		execMode = *req.ExecutionMode
	}
	sandboxStrat := ""
	if req.SandboxStrategy != nil {
		sandboxStrat = *req.SandboxStrategy
	}
	concPolicy := ""
	if req.ConcurrencyPolicy != nil {
		concPolicy = *req.ConcurrencyPolicy
	}
	if err := validateEnumFields(execMode, sandboxStrat, concPolicy); err != nil {
		return nil, err
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, workflow.ID, updates); err != nil {
			slog.ErrorContext(ctx, "failed to update workflow", "workflow_id", workflow.ID, "slug", slug, "org_id", orgID, "error", err)
			return nil, err
		}
		slog.InfoContext(ctx, "workflow updated", "workflow_id", workflow.ID, "slug", slug, "org_id", orgID)
	}

	return s.GetBySlug(ctx, orgID, slug)
}
