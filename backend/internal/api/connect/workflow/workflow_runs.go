package workflowconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	workflowv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/workflow/v1"
)

// WorkflowRunServiceInterface mirrors REST LoopHandler's workflowRunService dependency.
type WorkflowRunServiceInterface interface {
	ListWorkflowRuns(ctx context.Context, filter *workflowsvc.ListWorkflowRunsFilter) ([]*workflowDomain.WorkflowRun, int64, error)
	GetByID(ctx context.Context, id int64) (*workflowDomain.WorkflowRun, error)
	CountActiveRunsByWorkflowIDs(ctx context.Context, ids []int64) (map[int64]int64, error)
	GetAvgDuration(ctx context.Context, workflowID int64) (*float64, error)
}

// ListWorkflowRuns — REST analogue: GET /workflows/:slug/runs.
func (s *Server) ListWorkflowRuns(
	ctx context.Context, req *connect.Request[workflowv1.ListWorkflowRunsRequest],
) (*connect.Response[workflowv1.ListWorkflowRunsResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil || s.runSvc == nil {
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

	limit := int(req.Msg.GetLimit())
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := int(req.Msg.GetOffset())
	if offset < 0 {
		offset = 0
	}

	runs, total, err := s.runSvc.ListWorkflowRuns(ctx, &workflowsvc.ListWorkflowRunsFilter{
		WorkflowID: workflow.ID,
		Status:     req.Msg.GetStatus(),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*workflowv1.WorkflowRun, 0, len(runs))
	for _, r := range runs {
		items = append(items, toProtoWorkflowRun(r))
	}
	return connect.NewResponse(&workflowv1.ListWorkflowRunsResponse{
		Items:  items,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}), nil
}

// CancelWorkflowRun — REST analogue: POST /workflows/:slug/runs/:run_id/cancel.
func (s *Server) CancelWorkflowRun(
	ctx context.Context, req *connect.Request[workflowv1.CancelWorkflowRunRequest],
) (*connect.Response[workflowv1.CancelWorkflowRunResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil || s.runSvc == nil || s.orchestrator == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow services not configured"))
	}
	if req.Msg.GetWorkflowSlug() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("workflow_slug is required"))
	}
	if req.Msg.GetRunId() == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("run_id is required"))
	}
	tenant := middleware.GetTenant(ctx)

	workflow, err := s.svc.GetBySlug(ctx, tenant.OrganizationID, req.Msg.GetWorkflowSlug())
	if err != nil {
		if errors.Is(err, workflowsvc.ErrWorkflowNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	run, err := s.runSvc.GetByID(ctx, req.Msg.GetRunId())
	if err != nil {
		if errors.Is(err, workflowsvc.ErrRunNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("run not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if run.WorkflowID != workflow.ID {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("run not found"))
	}
	if run.IsTerminal() {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("run is already in terminal state"))
	}

	if run.PodKey != nil && s.podTerminator != nil {
		if err := s.podTerminator.TerminatePod(ctx, *run.PodKey); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	} else {
		if err := s.orchestrator.MarkRunCancelled(ctx, req.Msg.GetRunId(), "Cancelled by user"); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(&workflowv1.CancelWorkflowRunResponse{Message: "Run cancelled"}), nil
}
