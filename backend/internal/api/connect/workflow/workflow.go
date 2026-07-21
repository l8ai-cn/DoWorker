package workflowconnect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
)

const ServiceName = "proto.workflow.v1.WorkflowService"

const (
	ListWorkflowsProcedure     = "/" + ServiceName + "/ListWorkflows"
	GetWorkflowProcedure       = "/" + ServiceName + "/GetWorkflow"
	CreateWorkflowProcedure    = "/" + ServiceName + "/CreateWorkflow"
	UpdateWorkflowProcedure    = "/" + ServiceName + "/UpdateWorkflow"
	DeleteWorkflowProcedure    = "/" + ServiceName + "/DeleteWorkflow"
	EnableWorkflowProcedure    = "/" + ServiceName + "/EnableWorkflow"
	DisableWorkflowProcedure   = "/" + ServiceName + "/DisableWorkflow"
	TriggerWorkflowProcedure   = "/" + ServiceName + "/TriggerWorkflow"
	ListWorkflowRunsProcedure  = "/" + ServiceName + "/ListWorkflowRuns"
	CancelWorkflowRunProcedure = "/" + ServiceName + "/CancelWorkflowRun"
)

// WorkflowServiceInterface mirrors REST LoopHandler's workflowService dependency.
type WorkflowServiceInterface interface {
	List(ctx context.Context, filter *workflowsvc.ListWorkflowsFilter) ([]*workflowDomain.Workflow, int64, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*workflowDomain.Workflow, error)
	Create(ctx context.Context, req *workflowsvc.CreateWorkflowRequest) (*workflowDomain.Workflow, error)
	Update(ctx context.Context, orgID int64, slug string, req *workflowsvc.UpdateWorkflowRequest) (*workflowDomain.Workflow, error)
	Delete(ctx context.Context, orgID int64, slug string) error
	SetStatus(ctx context.Context, orgID int64, slug string, status string) (*workflowDomain.Workflow, error)
}

// PodTerminatorForWorkflow mirrors v1.PodTerminatorForWorkflow (ISP — only TerminatePod).
type PodTerminatorForWorkflow interface {
	TerminatePod(ctx context.Context, podKey string) error
}

type Server struct {
	svc           WorkflowServiceInterface
	runSvc        WorkflowRunServiceInterface
	orchestrator  WorkflowOrchestratorInterface
	orgSvc        middleware.OrganizationService
	podTerminator PodTerminatorForWorkflow
}

func NewServer(
	svc WorkflowServiceInterface,
	runSvc WorkflowRunServiceInterface,
	orchestrator WorkflowOrchestratorInterface,
	orgSvc middleware.OrganizationService,
	podTerminator PodTerminatorForWorkflow,
) *Server {
	return &Server{
		svc:           svc,
		runSvc:        runSvc,
		orchestrator:  orchestrator,
		orgSvc:        orgSvc,
		podTerminator: podTerminator,
	}
}

func Mount(mux *http.ServeMux, srv *Server, opts ...connect.HandlerOption) {
	mux.Handle(ListWorkflowsProcedure, connect.NewUnaryHandler(ListWorkflowsProcedure, srv.ListWorkflows, opts...))
	mux.Handle(GetWorkflowProcedure, connect.NewUnaryHandler(GetWorkflowProcedure, srv.GetWorkflow, opts...))
	mux.Handle(CreateWorkflowProcedure, connect.NewUnaryHandler(CreateWorkflowProcedure, srv.CreateWorkflow, opts...))
	mux.Handle(UpdateWorkflowProcedure, connect.NewUnaryHandler(UpdateWorkflowProcedure, srv.UpdateWorkflow, opts...))
	mux.Handle(DeleteWorkflowProcedure, connect.NewUnaryHandler(DeleteWorkflowProcedure, srv.DeleteWorkflow, opts...))
	mux.Handle(EnableWorkflowProcedure, connect.NewUnaryHandler(EnableWorkflowProcedure, srv.EnableWorkflow, opts...))
	mux.Handle(DisableWorkflowProcedure, connect.NewUnaryHandler(DisableWorkflowProcedure, srv.DisableWorkflow, opts...))
	mux.Handle(TriggerWorkflowProcedure, connect.NewUnaryHandler(TriggerWorkflowProcedure, srv.TriggerWorkflow, opts...))
	mux.Handle(ListWorkflowRunsProcedure, connect.NewUnaryHandler(ListWorkflowRunsProcedure, srv.ListWorkflowRuns, opts...))
	mux.Handle(CancelWorkflowRunProcedure, connect.NewUnaryHandler(CancelWorkflowRunProcedure, srv.CancelWorkflowRun, opts...))
}
