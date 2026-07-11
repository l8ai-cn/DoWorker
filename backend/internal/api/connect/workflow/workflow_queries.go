package workflowconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowsvc "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

func (s *Server) ListWorkflows(
	ctx context.Context, req *connect.Request[workflowv1.ListWorkflowsRequest],
) (*connect.Response[workflowv1.ListWorkflowsResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("workflow service not configured"))
	}
	tenant := middleware.GetTenant(ctx)
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
	var cronEnabled *bool
	if req.Msg.CronEnabled != nil {
		value := req.Msg.GetCronEnabled()
		cronEnabled = &value
	}
	workflows, total, err := s.svc.List(ctx, &workflowsvc.ListWorkflowsFilter{
		OrganizationID: tenant.OrganizationID,
		Status:         req.Msg.GetStatus(),
		ExecutionMode:  req.Msg.GetExecutionMode(),
		CronEnabled:    cronEnabled,
		Query:          req.Msg.GetQuery(),
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.addActiveRunCounts(ctx, workflows)
	items := make([]*workflowv1.Workflow, 0, len(workflows))
	for _, workflow := range workflows {
		items = append(items, toProtoWorkflow(workflow))
	}
	return connect.NewResponse(&workflowv1.ListWorkflowsResponse{
		Items:  items,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}), nil
}

func (s *Server) GetWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.GetWorkflowRequest],
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
	tenant := middleware.GetTenant(ctx)
	workflow, err := s.svc.GetBySlug(ctx, tenant.OrganizationID, req.Msg.GetWorkflowSlug())
	if err != nil {
		if errors.Is(err, workflowsvc.ErrWorkflowNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.addActiveRunCounts(ctx, []*workflowDomain.Workflow{workflow})
	if s.runSvc != nil {
		if avg, err := s.runSvc.GetAvgDuration(ctx, workflow.ID); err == nil && avg != nil {
			workflow.AvgDurationSec = avg
		}
	}
	return connect.NewResponse(toProtoWorkflow(workflow)), nil
}

func (s *Server) DeleteWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.DeleteWorkflowRequest],
) (*connect.Response[workflowv1.DeleteWorkflowResponse], error) {
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
	tenant := middleware.GetTenant(ctx)
	if err := s.svc.Delete(ctx, tenant.OrganizationID, req.Msg.GetWorkflowSlug()); err != nil {
		switch {
		case errors.Is(err, workflowsvc.ErrWorkflowNotFound):
			return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow not found"))
		case errors.Is(err, workflowsvc.ErrHasActiveRuns):
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workflow has active runs; cancel or wait first"))
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(&workflowv1.DeleteWorkflowResponse{Message: "Workflow deleted"}), nil
}

func (s *Server) addActiveRunCounts(ctx context.Context, workflows []*workflowDomain.Workflow) {
	if s.runSvc == nil || len(workflows) == 0 {
		return
	}
	ids := make([]int64, len(workflows))
	for i, workflow := range workflows {
		ids[i] = workflow.ID
	}
	counts, err := s.runSvc.CountActiveRunsByWorkflowIDs(ctx, ids)
	if err != nil {
		return
	}
	for _, workflow := range workflows {
		if count, ok := counts[workflow.ID]; ok {
			workflow.ActiveRunCount = int(count)
		}
	}
}
