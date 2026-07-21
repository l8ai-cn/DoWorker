package agentworkbenchconnect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	workbenchdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workbenchsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

const ServiceName = "proto.agent_workbench.v2.AgentWorkbenchService"

const (
	GetSessionSnapshotProcedure  = "/" + ServiceName + "/GetSessionSnapshot"
	StreamSessionDeltasProcedure = "/" + ServiceName + "/StreamSessionDeltas"
	ExecuteCommandProcedure      = "/" + ServiceName + "/ExecuteCommand"
)

type PersistenceRepository interface {
	workbenchdomain.PersistenceRepository
}

type SessionLookup interface {
	Get(context.Context, string) (*sessiondomain.Session, error)
}

type CommandExecutor interface {
	Execute(
		context.Context,
		*sessiondomain.Session,
		*agentworkbenchv2.CommandEnvelope,
	) (*agentworkbenchv2.CommandReceipt, error)
}

type Server struct {
	repository PersistenceRepository
	hub        *workbenchsvc.DeltaHub
	sessions   SessionLookup
	orgSvc     middleware.OrganizationService
	executor   CommandExecutor
}

func NewServer(
	repository PersistenceRepository,
	hub *workbenchsvc.DeltaHub,
	sessions SessionLookup,
	orgSvc middleware.OrganizationService,
	executor CommandExecutor,
) *Server {
	return &Server{
		repository: repository,
		hub:        hub,
		sessions:   sessions,
		orgSvc:     orgSvc,
		executor:   executor,
	}
}

func Mount(
	mux *http.ServeMux,
	server *Server,
	options ...connect.HandlerOption,
) {
	mux.Handle(GetSessionSnapshotProcedure, connect.NewUnaryHandler(
		GetSessionSnapshotProcedure,
		server.GetSessionSnapshot,
		options...,
	))
	mux.Handle(StreamSessionDeltasProcedure, connect.NewServerStreamHandler(
		StreamSessionDeltasProcedure,
		server.StreamSessionDeltas,
		options...,
	))
	mux.Handle(ExecuteCommandProcedure, connect.NewUnaryHandler(
		ExecuteCommandProcedure,
		server.ExecuteCommand,
		options...,
	))
}
