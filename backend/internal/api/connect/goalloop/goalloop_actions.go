package goalloopconnect

import (
	"context"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func (s *Server) CreateGoalLoop(
	ctx context.Context, req *connect.Request[goalloopv1.CreateGoalLoopRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.service == nil {
		return nil, unavailable()
	}
	tenant := middleware.GetTenant(ctx)
	loop, err := s.service.Create(ctx, goalloopsvc.CreateRequest{
		OrganizationID:       tenant.OrganizationID,
		CreatedByID:          tenant.UserID,
		Name:                 req.Msg.GetName(),
		Slug:                 req.Msg.GetSlug(),
		Description:          optionalString(req.Msg.GetDescription()),
		WorkerSpecSnapshotID: req.Msg.GetWorkerSpecSnapshotId(),
		Objective:            req.Msg.GetObjective(),
		AcceptanceCriteria:   req.Msg.GetAcceptanceCriteria(),
		VerificationCommand:  req.Msg.GetVerificationCommand(),
		MaxIterations:        optionalInt(req.Msg.MaxIterations),
		TokenBudget:          req.Msg.TokenBudget,
		TimeoutMinutes:       optionalInt(req.Msg.TimeoutMinutes),
		NoProgressLimit:      optionalInt(req.Msg.NoProgressLimit),
		SameErrorLimit:       optionalInt(req.Msg.SameErrorLimit),
		EscalationPolicy:     req.Msg.GetEscalationPolicy(),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	item, err := toProto(loop)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(item), nil
}

func (s *Server) StartGoalLoop(
	ctx context.Context, req *connect.Request[goalloopv1.GoalLoopActionRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	return s.action(ctx, req, func(ctx context.Context, orgID, userID int64, slug string) (*domain.GoalLoop, error) {
		return s.service.Start(ctx, orgID, userID, slug)
	})
}

func (s *Server) VerifyGoalLoop(
	ctx context.Context, req *connect.Request[goalloopv1.GoalLoopActionRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	return s.action(ctx, req, func(ctx context.Context, orgID, _ int64, slug string) (*domain.GoalLoop, error) {
		return s.service.Verify(ctx, orgID, slug)
	})
}

func (s *Server) CancelGoalLoop(
	ctx context.Context, req *connect.Request[goalloopv1.GoalLoopActionRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	return s.action(ctx, req, func(ctx context.Context, orgID, _ int64, slug string) (*domain.GoalLoop, error) {
		return s.service.Cancel(ctx, orgID, slug)
	})
}

type action func(context.Context, int64, int64, string) (*domain.GoalLoop, error)

func (s *Server) action(
	ctx context.Context,
	req *connect.Request[goalloopv1.GoalLoopActionRequest],
	run action,
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
	loop, err := run(ctx, tenant.OrganizationID, tenant.UserID, req.Msg.GetLoopSlug())
	if err != nil {
		return nil, mapServiceError(err)
	}
	item, err := toProto(loop)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(item), nil
}
