package grpc

import (
	"context"
	"errors"
	"strings"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
)

// mcpCreateWorkflow handles the "create_workflow" MCP method so pod agents (e.g. the
// looper guide conversation) can persist a Workflow after clarifying it with the
// user. Loops default to disabled — the lowest useful autonomy level — unless
// the caller explicitly passes enabled=true after user confirmation.
func (a *GRPCRunnerAdapter) mcpCreateWorkflow(ctx context.Context, tc *middleware.TenantContext, podKey string, payload []byte) (interface{}, *mcpError) {
	var params struct {
		Name               string `json:"name"`
		Description        string `json:"description"`
		PromptTemplate     string `json:"prompt_template"`
		AgentSlug          string `json:"agent_slug"`
		CronExpression     string `json:"cron_expression"`
		ExecutionMode      string `json:"execution_mode"`
		SandboxStrategy    string `json:"sandbox_strategy"`
		ConcurrencyPolicy  string `json:"concurrency_policy"`
		TimeoutMinutes     int    `json:"timeout_minutes"`
		MaxConcurrentRuns  int    `json:"max_concurrent_runs"`
		MaxRetainedRuns    int    `json:"max_retained_runs"`
		SessionPersistence bool   `json:"session_persistence"`
		RepositoryID       *int64 `json:"repository_id"`
		Enabled            bool   `json:"enabled"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}

	params.Name = strings.TrimSpace(params.Name)
	params.PromptTemplate = strings.TrimSpace(params.PromptTemplate)
	if params.Name == "" {
		return nil, newMcpError(400, "name is required")
	}
	if params.PromptTemplate == "" {
		return nil, newMcpError(400, "prompt_template is required")
	}
	if a.workflowService == nil {
		return nil, newMcpError(500, "workflow service not available")
	}

	req := &workflowService.CreateWorkflowRequest{
		OrganizationID:     tc.OrganizationID,
		CreatedByID:        tc.UserID,
		Name:               params.Name,
		AgentSlug:          params.AgentSlug,
		PromptTemplate:     params.PromptTemplate,
		RepositoryID:       params.RepositoryID,
		ExecutionMode:      params.ExecutionMode,
		SandboxStrategy:    params.SandboxStrategy,
		ConcurrencyPolicy:  params.ConcurrencyPolicy,
		TimeoutMinutes:     params.TimeoutMinutes,
		MaxConcurrentRuns:  params.MaxConcurrentRuns,
		MaxRetainedRuns:    params.MaxRetainedRuns,
		SessionPersistence: params.SessionPersistence,
	}
	if desc := strings.TrimSpace(params.Description); desc != "" {
		req.Description = &desc
	}
	if cron := strings.TrimSpace(params.CronExpression); cron != "" {
		req.CronExpression = &cron
	}
	a.applyCallingPodDefaults(ctx, podKey, req)
	if req.AgentSlug == "" {
		return nil, newMcpError(400, "agent_slug is required (could not infer from calling pod)")
	}

	workflow, err := a.workflowService.Create(ctx, req)
	if err != nil {
		return nil, mapWorkflowCreateError(err)
	}

	if !params.Enabled {
		if updated, statusErr := a.workflowService.SetStatus(ctx, tc.OrganizationID, workflow.Slug, workflowDomain.StatusDisabled); statusErr == nil {
			workflow = updated
		}
	}

	return map[string]interface{}{
		"workflow": toMCPWorkflowSummary(workflow),
	}, nil
}

// applyCallingPodDefaults inherits agent/runner from the pod that hosts the
// guiding agent, so a conversational create works without the user having to
// know infrastructure IDs.
func (a *GRPCRunnerAdapter) applyCallingPodDefaults(ctx context.Context, podKey string, req *workflowService.CreateWorkflowRequest) {
	if a.podService == nil || podKey == "" {
		return
	}
	pod, err := a.podService.GetPodByKey(ctx, podKey)
	if err != nil || pod == nil {
		return
	}
	if req.AgentSlug == "" {
		req.AgentSlug = pod.AgentSlug
	}
	if req.RunnerID == nil && req.AgentSlug == pod.AgentSlug && pod.RunnerID != 0 {
		runnerID := pod.RunnerID
		req.RunnerID = &runnerID
	}
}

func mapWorkflowCreateError(err error) *mcpError {
	switch {
	case errors.Is(err, workflowService.ErrInvalidCron):
		return newMcpError(400, "invalid cron expression (standard 5-field format: minute hour day month weekday)")
	case errors.Is(err, workflowService.ErrInvalidSlug):
		return newMcpError(400, "workflow name cannot be converted to a valid slug")
	case errors.Is(err, workflowService.ErrDuplicateSlug):
		return newMcpError(409, "a workflow with this name already exists")
	case errors.Is(err, workflowService.ErrInvalidEnumValue):
		return newMcpError(400, err.Error())
	default:
		return newMcpErrorf(500, "failed to create workflow: %v", err)
	}
}
