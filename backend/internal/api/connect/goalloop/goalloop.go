package goalloopconnect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	workerspecdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	goalloopsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/goalloop"
)

const ServiceName = "proto.goalloop.v1.GoalLoopService"

const (
	ListWorkerSnapshotsProcedure = "/" + ServiceName + "/ListWorkerSnapshots"
	CompileLoopProgramProcedure  = "/" + ServiceName + "/CompileLoopProgram"
	GenerateLoopProgramProcedure = "/" + ServiceName + "/GenerateLoopProgram"
	RepairLoopProgramProcedure   = "/" + ServiceName + "/RepairLoopProgram"
	RunLoopProgramProcedure      = "/" + ServiceName + "/RunLoopProgram"
	ListGoalLoopsProcedure       = "/" + ServiceName + "/ListGoalLoops"
	GetGoalLoopProcedure         = "/" + ServiceName + "/GetGoalLoop"
	CreateGoalLoopProcedure      = "/" + ServiceName + "/CreateGoalLoop"
	StartGoalLoopProcedure       = "/" + ServiceName + "/StartGoalLoop"
	VerifyGoalLoopProcedure      = "/" + ServiceName + "/VerifyGoalLoop"
	CancelGoalLoopProcedure      = "/" + ServiceName + "/CancelGoalLoop"
)

type GoalLoopService interface {
	ListWorkerSnapshots(
		ctx context.Context,
		organizationID, userID int64,
	) ([]workerspecdomain.Snapshot, error)
	ValidateWorkerSnapshotForExecution(
		ctx context.Context,
		organizationID, userID, snapshotID int64,
	) error
	ValidateExecutionReady() error
	List(ctx context.Context, filter domain.ListFilter) ([]*domain.GoalLoop, int64, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
	Create(ctx context.Context, req goalloopsvc.CreateRequest) (*domain.GoalLoop, error)
	Start(ctx context.Context, orgID, userID int64, slug string) (*domain.GoalLoop, error)
	Verify(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
	Cancel(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error)
}

type LoopDraftGenerator interface {
	Generate(
		context.Context,
		goalloopsvc.DraftGenerationScope,
		goalloopsvc.DraftGenerationInput,
	) (goalloopsvc.DraftProposal, error)
	Repair(
		context.Context,
		goalloopsvc.DraftGenerationScope,
		goalloopsvc.DraftRepairInput,
	) (goalloopsvc.DraftRepairProposal, error)
}

type Server struct {
	service  GoalLoopService
	orgSvc   middleware.OrganizationService
	aiDrafts LoopDraftGenerator
}

type Option func(*Server)

func WithAIGeneration(generator LoopDraftGenerator) Option {
	return func(server *Server) {
		server.aiDrafts = generator
	}
}

func NewServer(
	service GoalLoopService,
	orgSvc middleware.OrganizationService,
	options ...Option,
) *Server {
	server := &Server{service: service, orgSvc: orgSvc}
	for _, option := range options {
		option(server)
	}
	return server
}

func Mount(mux *http.ServeMux, srv *Server, opts ...connect.HandlerOption) {
	mux.Handle(ListWorkerSnapshotsProcedure, connect.NewUnaryHandler(ListWorkerSnapshotsProcedure, srv.ListWorkerSnapshots, opts...))
	mux.Handle(CompileLoopProgramProcedure, connect.NewUnaryHandler(CompileLoopProgramProcedure, srv.CompileLoopProgram, opts...))
	mux.Handle(GenerateLoopProgramProcedure, connect.NewUnaryHandler(GenerateLoopProgramProcedure, srv.GenerateLoopProgram, opts...))
	mux.Handle(RepairLoopProgramProcedure, connect.NewUnaryHandler(RepairLoopProgramProcedure, srv.RepairLoopProgram, opts...))
	mux.Handle(RunLoopProgramProcedure, connect.NewUnaryHandler(RunLoopProgramProcedure, srv.RunLoopProgram, opts...))
	mux.Handle(ListGoalLoopsProcedure, connect.NewUnaryHandler(ListGoalLoopsProcedure, srv.ListGoalLoops, opts...))
	mux.Handle(GetGoalLoopProcedure, connect.NewUnaryHandler(GetGoalLoopProcedure, srv.GetGoalLoop, opts...))
	mux.Handle(CreateGoalLoopProcedure, connect.NewUnaryHandler(CreateGoalLoopProcedure, srv.CreateGoalLoop, opts...))
	mux.Handle(StartGoalLoopProcedure, connect.NewUnaryHandler(StartGoalLoopProcedure, srv.StartGoalLoop, opts...))
	mux.Handle(VerifyGoalLoopProcedure, connect.NewUnaryHandler(VerifyGoalLoopProcedure, srv.VerifyGoalLoop, opts...))
	mux.Handle(CancelGoalLoopProcedure, connect.NewUnaryHandler(CancelGoalLoopProcedure, srv.CancelGoalLoop, opts...))
}
