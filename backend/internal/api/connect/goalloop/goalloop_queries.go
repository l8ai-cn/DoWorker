package goalloopconnect

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	goalloopsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/goalloop"
	goalloopv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/goalloop/v1"
)

func (s *Server) ListWorkerSnapshots(
	ctx context.Context, req *connect.Request[goalloopv1.ListWorkerSnapshotsRequest],
) (*connect.Response[goalloopv1.ListWorkerSnapshotsResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.service == nil {
		return nil, unavailable()
	}
	tenant := middleware.GetTenant(ctx)
	snapshots, err := s.service.ListWorkerSnapshots(
		ctx,
		tenant.OrganizationID,
		tenant.UserID,
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	items := make([]*goalloopv1.WorkerSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		items = append(items, &goalloopv1.WorkerSnapshot{
			Id:         snapshot.ID,
			Alias:      snapshot.Summary.Alias,
			WorkerType: string(snapshot.Summary.WorkerType.Slug),
			CreatedAt:  snapshot.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return connect.NewResponse(&goalloopv1.ListWorkerSnapshotsResponse{Items: items}), nil
}

func (s *Server) ListGoalLoops(
	ctx context.Context, req *connect.Request[goalloopv1.ListGoalLoopsRequest],
) (*connect.Response[goalloopv1.ListGoalLoopsResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.service == nil {
		return nil, unavailable()
	}
	limit, offset := page(req.Msg.GetLimit(), req.Msg.GetOffset())
	tenant := middleware.GetTenant(ctx)
	loops, total, err := s.service.List(ctx, goalLoopListFilter(
		tenant.OrganizationID,
		req.Msg.GetStatus(),
		req.Msg.GetQuery(),
		limit,
		offset,
	))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*goalloopv1.GoalLoop, 0, len(loops))
	for _, loop := range loops {
		item, err := toProto(loop)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		items = append(items, item)
	}
	return connect.NewResponse(&goalloopv1.ListGoalLoopsResponse{
		Items:  items,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}), nil
}

func (s *Server) GetGoalLoop(
	ctx context.Context, req *connect.Request[goalloopv1.GetGoalLoopRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.service == nil {
		return nil, unavailable()
	}
	if req.Msg.GetLoopSlug() == "" {
		return nil, invalid("loop_slug is required")
	}
	tenant := middleware.GetTenant(ctx)
	loop, err := s.service.GetBySlug(ctx, tenant.OrganizationID, req.Msg.GetLoopSlug())
	if err != nil {
		return nil, mapServiceError(err)
	}
	item, err := toProto(loop)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(item), nil
}

func goalLoopListFilter(orgID int64, status, query string, limit, offset int) domain.ListFilter {
	return domain.ListFilter{
		OrganizationID: orgID,
		Status:         status,
		Query:          query,
		Limit:          limit,
		Offset:         offset,
	}
}

func page(limit, offset int32) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return int(limit), int(offset)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, goalloopsvc.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("goal loop not found"))
	case errors.Is(err, goalloopsvc.ErrInvalidInput):
		return invalid("invalid goal loop input")
	case errors.Is(err, goalloopsvc.ErrInvalidState), errors.Is(err, goalloopsvc.ErrVerificationPending):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, goalloopsvc.ErrExecutionUnavailable):
		return unavailable()
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func invalid(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(message))
}

func unavailable() error {
	return connect.NewError(connect.CodeUnavailable, errors.New("goal loop service not configured"))
}
