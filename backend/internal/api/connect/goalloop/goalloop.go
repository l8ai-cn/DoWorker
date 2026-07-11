package goalloopconnect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
)

const ServiceName = "proto.goalloop.v1.GoalLoopService"

const (
	ListGoalLoopsProcedure  = "/" + ServiceName + "/ListGoalLoops"
	GetGoalLoopProcedure    = "/" + ServiceName + "/GetGoalLoop"
	CreateGoalLoopProcedure = "/" + ServiceName + "/CreateGoalLoop"
	StartGoalLoopProcedure  = "/" + ServiceName + "/StartGoalLoop"
	VerifyGoalLoopProcedure = "/" + ServiceName + "/VerifyGoalLoop"
	CancelGoalLoopProcedure = "/" + ServiceName + "/CancelGoalLoop"
)

type GoalLoopService interface {
	List(ctx context.Context, filter domain.ListFilter) ([]*domain.GoalLoop, int64, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
	Create(ctx context.Context, req goalloopsvc.CreateRequest) (*domain.GoalLoop, error)
	Start(ctx context.Context, orgID, userID int64, slug string) (*domain.GoalLoop, error)
	Verify(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
	Cancel(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
}

type Server struct {
	service GoalLoopService
	orgSvc  middleware.OrganizationService
}

func NewServer(service GoalLoopService, orgSvc middleware.OrganizationService) *Server {
	return &Server{service: service, orgSvc: orgSvc}
}

func Mount(mux *http.ServeMux, srv *Server, opts ...connect.HandlerOption) {
	mux.Handle(ListGoalLoopsProcedure, connect.NewUnaryHandler(ListGoalLoopsProcedure, srv.ListGoalLoops, opts...))
	mux.Handle(GetGoalLoopProcedure, connect.NewUnaryHandler(GetGoalLoopProcedure, srv.GetGoalLoop, opts...))
	mux.Handle(CreateGoalLoopProcedure, connect.NewUnaryHandler(CreateGoalLoopProcedure, srv.CreateGoalLoop, opts...))
	mux.Handle(StartGoalLoopProcedure, connect.NewUnaryHandler(StartGoalLoopProcedure, srv.StartGoalLoop, opts...))
	mux.Handle(VerifyGoalLoopProcedure, connect.NewUnaryHandler(VerifyGoalLoopProcedure, srv.VerifyGoalLoop, opts...))
	mux.Handle(CancelGoalLoopProcedure, connect.NewUnaryHandler(CancelGoalLoopProcedure, srv.CancelGoalLoop, opts...))
}
