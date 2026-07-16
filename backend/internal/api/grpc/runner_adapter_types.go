package grpc

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/anthropics/agentsmesh/backend/internal/infra/pki"
	"github.com/anthropics/agentsmesh/backend/internal/interfaces"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/binding"
	blockstoreservice "github.com/anthropics/agentsmesh/backend/internal/service/blockstore"
	"github.com/anthropics/agentsmesh/backend/internal/service/channel"
	goalloopService "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	knowledgebaseservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// PodEventSink forwards runner pod events into agent session streaming.
type PodEventSink interface {
	HandleAcpSession(ctx context.Context, podKey, eventType, payloadJSON string)
	PublishPodStatus(ctx context.Context, podKey, podStatus, agentStatus string)
	HandlePodUsage(ctx context.Context, evt *runnerv1.PodUsageEvent)
	UpdateExternalSessionID(ctx context.Context, podKey, externalID string)
}

type WorkbenchEventSink interface {
	HandleWorkbenchEvents(
		ctx context.Context,
		runnerID int64,
		batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
	) error
}

const certRevocationCheckInterval = 5 * time.Minute

var _ runnerv1.RunnerServiceServer = (*GRPCRunnerAdapter)(nil)

type GRPCRunnerAdapter struct {
	runnerv1.UnimplementedRunnerServiceServer

	logger         *slog.Logger
	db             *gorm.DB
	runnerService  RunnerServiceInterface
	orgService     OrganizationServiceInterface
	pkiService     *pki.Service
	agentsProvider interfaces.AgentsProvider

	connManager  *runner.RunnerConnectionManager
	readyResults *runnerReadyResultTracker

	podEvents       PodEventSink
	workbenchEvents WorkbenchEventSink

	podService           *agentpod.PodService
	mcpPodService        *agentpod.PodService
	podOrchestrator      *agentpod.PodOrchestrator
	channelService       *channel.Service
	bindingService       *binding.Service
	ticketService        *ticket.Service
	repositoryService    repository.RepositoryServiceInterface
	runnerMcpService     *runner.Service
	agentSvc             *agent.AgentService
	userConfigSvc        *agent.UserConfigService
	podRouter            PodRouterForMCP // *runner.PodRouter, optional
	workflowService      *workflowService.WorkflowService
	workflowRunService   *workflowService.WorkflowRunService
	workflowOrchestrator *workflowService.WorkflowOrchestrator
	goalLoopService      *goalloopService.Service
	blockstoreService    *blockstoreservice.Service
	knowledgebaseService *knowledgebaseservice.Service
}

type MCPDependencies struct {
	PodService           *agentpod.PodService
	PodOrchestrator      *agentpod.PodOrchestrator
	ChannelService       *channel.Service
	BindingService       *binding.Service
	TicketService        *ticket.Service
	RepositoryService    repository.RepositoryServiceInterface
	RunnerService        *runner.Service
	AgentSvc             *agent.AgentService
	UserConfigSvc        *agent.UserConfigService
	PodRouter            PodRouterForMCP // *runner.PodRouter, optional
	WorkflowService      *workflowService.WorkflowService
	WorkflowRunService   *workflowService.WorkflowRunService
	WorkflowOrchestrator *workflowService.WorkflowOrchestrator
	GoalLoopService      *goalloopService.Service
	BlockstoreService    *blockstoreservice.Service
	KnowledgebaseService *knowledgebaseservice.Service
}

func NewGRPCRunnerAdapter(
	logger *slog.Logger,
	db *gorm.DB,
	runnerService RunnerServiceInterface,
	orgService OrganizationServiceInterface,
	pkiService *pki.Service,
	agentsProvider interfaces.AgentsProvider,
	connManager *runner.RunnerConnectionManager,
	mcpDeps *MCPDependencies,
) *GRPCRunnerAdapter {
	adapter := &GRPCRunnerAdapter{
		logger:         logger,
		db:             db,
		runnerService:  runnerService,
		orgService:     orgService,
		pkiService:     pkiService,
		agentsProvider: agentsProvider,
		connManager:    connManager,
		readyResults:   newRunnerReadyResultTracker(),
	}

	if mcpDeps != nil {
		adapter.podService = mcpDeps.PodService
		adapter.mcpPodService = mcpDeps.PodService
		adapter.podOrchestrator = mcpDeps.PodOrchestrator
		adapter.channelService = mcpDeps.ChannelService
		adapter.bindingService = mcpDeps.BindingService
		adapter.ticketService = mcpDeps.TicketService
		adapter.repositoryService = mcpDeps.RepositoryService
		adapter.runnerMcpService = mcpDeps.RunnerService
		adapter.agentSvc = mcpDeps.AgentSvc
		adapter.userConfigSvc = mcpDeps.UserConfigSvc
		adapter.podRouter = mcpDeps.PodRouter
		adapter.workflowService = mcpDeps.WorkflowService
		adapter.workflowRunService = mcpDeps.WorkflowRunService
		adapter.workflowOrchestrator = mcpDeps.WorkflowOrchestrator
		adapter.goalLoopService = mcpDeps.GoalLoopService
		adapter.blockstoreService = mcpDeps.BlockstoreService
		adapter.knowledgebaseService = mcpDeps.KnowledgebaseService
	}

	return adapter
}
