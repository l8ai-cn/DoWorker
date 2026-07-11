package workflowconnect

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowsvc "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

// WorkflowOrchestratorInterface mirrors REST LoopHandler's orchestrator dependency.
type WorkflowOrchestratorInterface interface {
	TriggerRun(ctx context.Context, req *workflowsvc.TriggerRunRequest) (*workflowsvc.TriggerRunResult, error)
	StartRun(ctx context.Context, workflow *workflowDomain.Workflow, run *workflowDomain.WorkflowRun, userID int64)
	MarkRunCancelled(ctx context.Context, runID int64, reason string) error
}

// EnableWorkflow — REST analogue: POST /workflows/:slug/enable.
func (s *Server) EnableWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.WorkflowActionRequest],
) (*connect.Response[workflowv1.Workflow], error) {
	return s.setStatus(ctx, req.Msg, workflowDomain.StatusEnabled)
}

// DisableWorkflow — REST analogue: POST /workflows/:slug/disable.
func (s *Server) DisableWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.WorkflowActionRequest],
) (*connect.Response[workflowv1.Workflow], error) {
	return s.setStatus(ctx, req.Msg, workflowDomain.StatusDisabled)
}

func (s *Server) setStatus(
	ctx context.Context, m *workflowv1.WorkflowActionRequest, status string,
) (*connect.Response[workflowv1.Workflow], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, m, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow service not configured"))
	}
	if m.GetWorkflowSlug() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("workflow_slug is required"))
	}
	tenant := middleware.GetTenant(ctx)
	workflow, err := s.svc.SetStatus(ctx, tenant.OrganizationID, m.GetWorkflowSlug(), status)
	if err != nil {
		if errors.Is(err, workflowsvc.ErrWorkflowNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(toProtoWorkflow(workflow)), nil
}

// TriggerWorkflow — REST analogue: POST /workflows/:slug/trigger.
func (s *Server) TriggerWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.TriggerWorkflowRequest],
) (*connect.Response[workflowv1.TriggerWorkflowResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil || s.orchestrator == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow services not configured"))
	}
	if req.Msg.GetWorkflowSlug() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("workflow_slug is required"))
	}
	tenant := middleware.GetTenant(ctx)
	workflow, err := s.svc.GetBySlug(ctx, tenant.OrganizationID, req.Msg.GetWorkflowSlug())
	if err != nil {
		if errors.Is(err, workflowsvc.ErrWorkflowNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	var variables json.RawMessage
	if v := req.Msg.GetVariablesJson(); v != "" {
		variables = json.RawMessage(v)
	}
	result, err := s.orchestrator.TriggerRun(ctx, &workflowsvc.TriggerRunRequest{
		WorkflowID:    workflow.ID,
		TriggerType:   workflowDomain.RunTriggerManual,
		TriggerSource: "user:" + strconv.FormatInt(tenant.UserID, 10),
		TriggerParams: variables,
	})
	if err != nil {
		if errors.Is(err, workflowsvc.ErrWorkflowDisabled) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workflow is disabled"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if result.Skipped {
		return connect.NewResponse(&workflowv1.TriggerWorkflowResponse{
			Run:     toProtoWorkflowRun(result.Run),
			Skipped: true,
			Reason:  result.Reason,
		}), nil
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	go func() {
		defer startCancel()
		s.orchestrator.StartRun(startCtx, result.Workflow, result.Run, tenant.UserID)
	}()

	return connect.NewResponse(&workflowv1.TriggerWorkflowResponse{
		Run:     toProtoWorkflowRun(result.Run),
		Skipped: false,
	}), nil
}
