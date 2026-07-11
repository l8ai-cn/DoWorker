package workflow

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	ticketSvc "github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type PodTerminator interface {
	TerminatePod(ctx context.Context, podKey string) error
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

	podOrchestrator *agentpodSvc.PodOrchestrator
	autopilotSvc    *agentpodSvc.AutopilotControllerService
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
	podOrch *agentpodSvc.PodOrchestrator,
	autopilot *agentpodSvc.AutopilotControllerService,
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
