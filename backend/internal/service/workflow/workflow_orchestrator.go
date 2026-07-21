package workflow

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	agentpodSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	ticketSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/ticket"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type PodTerminator interface {
	TerminatePod(ctx context.Context, podKey string) error
}

type PodCreator interface {
	CreatePod(
		ctx context.Context,
		req *agentpodSvc.OrchestrateCreatePodRequest,
	) (*agentpodSvc.OrchestrateCreatePodResult, error)
}

type AutopilotStarter interface {
	CreateAndStart(
		ctx context.Context,
		req *agentpodSvc.CreateAndStartRequest,
	) (*agentpod.AutopilotController, error)
	GetApprovalTimedOut(
		ctx context.Context,
		orgIDs []int64,
	) ([]*agentpod.AutopilotController, error)
	UpdateAutopilotControllerStatus(
		ctx context.Context,
		autopilotKey string,
		updates map[string]interface{},
	) error
}

type RepoQueryForWorkflow interface {
	GetByID(ctx context.Context, id int64) (*gitprovider.Repository, error)
}

// WorkflowOrchestrator never owns run.Status — Pod is SSOT, status is derived on read.
type WorkflowOrchestrator struct {
	workflowService    *WorkflowService
	workflowRunService *WorkflowRunService
	eventBus           *eventbus.EventBus
	logger             *slog.Logger

	podOrchestrator PodCreator
	autopilotSvc    AutopilotStarter
	podTerminator   PodTerminator
	ticketService   *ticketSvc.Service
	repoQuery       RepoQueryForWorkflow

	httpClient *http.Client
}

func NewWorkflowOrchestrator(
	workflowService *WorkflowService,
	workflowRunService *WorkflowRunService,
	eventBus *eventbus.EventBus,
	logger *slog.Logger,
) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		workflowService:    workflowService,
		workflowRunService: workflowRunService,
		eventBus:           eventBus,
		logger:             logger.With("component", "loop_orchestrator"),
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (o *WorkflowOrchestrator) SetPodDependencies(
	podOrch PodCreator,
	autopilot AutopilotStarter,
	podTerminator PodTerminator,
	ticket *ticketSvc.Service,
	repoQuery RepoQueryForWorkflow,
) {
	o.podOrchestrator = podOrch
	o.autopilotSvc = autopilot
	o.podTerminator = podTerminator
	o.ticketService = ticket
	o.repoQuery = repoQuery
}
