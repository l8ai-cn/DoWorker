package workflowconnect

import (
	"context"
	"encoding/json"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowsvc "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

func jsonRawFromString(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}

// CreateWorkflow — REST analogue: POST /workflows.
func (s *Server) CreateWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.CreateWorkflowRequest],
) (*connect.Response[workflowv1.Workflow], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow service not configured"))
	}
	tenant := middleware.GetTenant(ctx)
	m := req.Msg

	maxConcurrent := 1
	if v := m.GetMaxConcurrentRuns(); m.MaxConcurrentRuns != nil {
		maxConcurrent = int(v)
	}
	if maxConcurrent < 1 || maxConcurrent > 10 {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("max_concurrent_runs must be between 1 and 10"))
	}
	maxRetained := 0
	if m.MaxRetainedRuns != nil {
		maxRetained = int(m.GetMaxRetainedRuns())
	}
	if maxRetained < 0 || maxRetained > 10000 {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("max_retained_runs must be between 0 and 10000"))
	}
	timeoutMin := 60
	if m.TimeoutMinutes != nil {
		timeoutMin = int(m.GetTimeoutMinutes())
	}
	if timeoutMin < 1 || timeoutMin > 1440 {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("timeout_minutes must be between 1 and 1440"))
	}
	sessionPersist := true
	if m.SessionPersistence != nil {
		sessionPersist = m.GetSessionPersistence()
	}
	idleTimeout := 30
	if m.IdleTimeoutSec != nil {
		idleTimeout = int(m.GetIdleTimeoutSec())
	}
	if idleTimeout < 0 || idleTimeout > 3600 {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("idle_timeout_sec must be between 0 and 3600"))
	}

	svcReq := &workflowsvc.CreateWorkflowRequest{
		OrganizationID:     tenant.OrganizationID,
		CreatedByID:        tenant.UserID,
		Name:               m.GetName(),
		Slug:               m.GetSlug(),
		AgentSlug:          m.GetAgentSlug(),
		PermissionMode:     m.GetPermissionMode(),
		PromptTemplate:     m.GetPromptTemplate(),
		PromptVariables:    jsonRawFromString(m.GetPromptVariablesJson()),
		ConfigOverrides:    jsonRawFromString(m.GetConfigOverridesJson()),
		AutopilotConfig:    jsonRawFromString(m.GetAutopilotConfigJson()),
		RepositoryID:       m.RepositoryId,
		RunnerID:           m.RunnerId,
		TicketID:           m.TicketId,
		ModelResourceID:    m.ModelResourceId,
		UsedEnvBundles:     m.GetUsedEnvBundles(),
		ExecutionMode:      m.GetExecutionMode(),
		SandboxStrategy:    m.GetSandboxStrategy(),
		SessionPersistence: sessionPersist,
		ConcurrencyPolicy:  m.GetConcurrencyPolicy(),
		MaxConcurrentRuns:  maxConcurrent,
		MaxRetainedRuns:    maxRetained,
		TimeoutMinutes:     timeoutMin,
		IdleTimeoutSec:     idleTimeout,
	}
	if v := m.GetDescription(); v != "" {
		svcReq.Description = &v
	}
	if v := m.GetBranchName(); v != "" {
		svcReq.BranchName = &v
	}
	if v := m.GetCronExpression(); v != "" {
		svcReq.CronExpression = &v
	}
	if v := m.GetCallbackUrl(); v != "" {
		svcReq.CallbackURL = &v
	}

	workflow, err := s.svc.Create(ctx, svcReq)
	if err != nil {
		switch {
		case errors.Is(err, workflowsvc.ErrDuplicateSlug):
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("workflow slug already exists"))
		case errors.Is(err, workflowsvc.ErrInvalidSlug):
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid slug format"))
		case errors.Is(err, workflowsvc.ErrInvalidCron):
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid cron expression"))
		case errors.Is(err, workflowsvc.ErrInvalidCallbackURL),
			errors.Is(err, workflowsvc.ErrInvalidEnumValue):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(toProtoWorkflow(workflow)), nil
}
