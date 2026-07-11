package workflowconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowsvc "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

func (s *Server) UpdateWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.UpdateWorkflowRequest],
) (*connect.Response[workflowv1.Workflow], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow service not configured"))
	}
	if req.Msg.GetWorkflowSlug() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("workflow_slug is required"))
	}
	if err := validateUpdateBounds(req.Msg); err != nil {
		return nil, err
	}
	tenant := middleware.GetTenant(ctx)
	workflow, err := s.svc.Update(ctx, tenant.OrganizationID, req.Msg.GetWorkflowSlug(), buildUpdateRequest(req.Msg))
	if err != nil {
		switch {
		case errors.Is(err, workflowsvc.ErrWorkflowNotFound):
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		case errors.Is(err, workflowsvc.ErrInvalidCron):
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid cron expression"))
		case errors.Is(err, workflowsvc.ErrInvalidCallbackURL), errors.Is(err, workflowsvc.ErrInvalidEnumValue):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(toProtoWorkflow(workflow)), nil
}

func validateUpdateBounds(m *workflowv1.UpdateWorkflowRequest) error {
	if m.MaxConcurrentRuns != nil && (m.GetMaxConcurrentRuns() < 1 || m.GetMaxConcurrentRuns() > 10) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("max_concurrent_runs must be between 1 and 10"))
	}
	if m.TimeoutMinutes != nil && (m.GetTimeoutMinutes() < 1 || m.GetTimeoutMinutes() > 1440) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("timeout_minutes must be between 1 and 1440"))
	}
	if m.MaxRetainedRuns != nil && (m.GetMaxRetainedRuns() < 0 || m.GetMaxRetainedRuns() > 10000) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("max_retained_runs must be between 0 and 10000"))
	}
	if m.IdleTimeoutSec != nil && (m.GetIdleTimeoutSec() < 0 || m.GetIdleTimeoutSec() > 3600) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("idle_timeout_sec must be between 0 and 3600"))
	}
	return nil
}

func buildUpdateRequest(m *workflowv1.UpdateWorkflowRequest) *workflowsvc.UpdateWorkflowRequest {
	request := &workflowsvc.UpdateWorkflowRequest{
		AgentSlug:          m.GetAgentSlug(),
		Name:               m.Name,
		Description:        m.Description,
		PermissionMode:     m.PermissionMode,
		PromptTemplate:     m.PromptTemplate,
		ExecutionMode:      m.ExecutionMode,
		CronExpression:     m.CronExpression,
		CallbackURL:        m.CallbackUrl,
		SandboxStrategy:    m.SandboxStrategy,
		SessionPersistence: m.SessionPersistence,
		ConcurrencyPolicy:  m.ConcurrencyPolicy,
		BranchName:         m.BranchName,
		RepositoryID:       m.RepositoryId,
		RunnerID:           m.RunnerId,
		TicketID:           m.TicketId,
		ModelResourceID:    m.ModelResourceId,
	}
	if value := m.GetPromptVariablesJson(); value != "" {
		request.PromptVariables = jsonRawFromString(value)
	}
	if value := m.GetConfigOverridesJson(); value != "" {
		request.ConfigOverrides = jsonRawFromString(value)
	}
	if value := m.GetAutopilotConfigJson(); value != "" {
		request.AutopilotConfig = jsonRawFromString(value)
	}
	if m.MaxConcurrentRuns != nil {
		value := int(m.GetMaxConcurrentRuns())
		request.MaxConcurrentRuns = &value
	}
	if m.MaxRetainedRuns != nil {
		value := int(m.GetMaxRetainedRuns())
		request.MaxRetainedRuns = &value
	}
	if m.TimeoutMinutes != nil {
		value := int(m.GetTimeoutMinutes())
		request.TimeoutMinutes = &value
	}
	if m.IdleTimeoutSec != nil {
		value := int(m.GetIdleTimeoutSec())
		request.IdleTimeoutSec = &value
	}
	if m.UsedEnvBundles != nil {
		names := m.GetUsedEnvBundles().GetNames()
		if names == nil {
			names = []string{}
		}
		request.UsedEnvBundles = &names
	}
	return request
}
