package v1

import (
	grpcserver "github.com/anthropics/agentsmesh/backend/internal/api/grpc"
	agentworkbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/anthropics/agentsmesh/backend/internal/infra/acme"
	"github.com/anthropics/agentsmesh/backend/internal/infra/email"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/infra/websocket"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	agentworkbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	apikeyservice "github.com/anthropics/agentsmesh/backend/internal/service/apikey"
	"github.com/anthropics/agentsmesh/backend/internal/service/auth"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/anthropics/agentsmesh/backend/internal/service/channel"
	coordinatorservice "github.com/anthropics/agentsmesh/backend/internal/service/coordinator"
	envbundlesvc "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	expertSvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	"github.com/anthropics/agentsmesh/backend/internal/service/geo"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	imbridgesvc "github.com/anthropics/agentsmesh/backend/internal/service/imbridge"
	"github.com/anthropics/agentsmesh/backend/internal/service/invitation"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	previewservice "github.com/anthropics/agentsmesh/backend/internal/service/preview"
	"github.com/anthropics/agentsmesh/backend/internal/service/promocode"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerlogservice "github.com/anthropics/agentsmesh/backend/internal/service/runnerlog"
	skillSvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	ssoservice "github.com/anthropics/agentsmesh/backend/internal/service/sso"
	supportticketservice "github.com/anthropics/agentsmesh/backend/internal/service/supportticket"
	"github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	tokenusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenusage"
	"github.com/anthropics/agentsmesh/backend/internal/service/user"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	workerspecsvc "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	workflow "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/redis/go-redis/v9"
)

// MessageService is a type alias for agent.MessageService
type MessageService = agent.MessageService

// Services holds all service dependencies for API handlers
type Services struct {
	Auth *auth.Service
	User *user.Service
	Org  *organization.Service
	// Agent services (split by responsibility)
	AgentSvc             *agent.AgentService
	UserConfig           *agent.UserConfigService
	Repository           *repository.Service
	Webhook              *repository.WebhookService // Webhook management for repositories
	Runner               *runner.Service
	RunnerConnMgr        *runner.RunnerConnectionManager // Runner gRPC connection manager
	PodCoordinator       *runner.PodCoordinator          // Pod lifecycle coordinator
	Pod                  *agentpod.PodService
	PodOrchestrator      *agentpod.PodOrchestrator // Unified Pod creation orchestrator
	WorkerCreation       *workercreation.Service
	WorkerPlanAuthorizer QuickTaskPlanAuthorizer
	WorkerApply          QuickTaskPlanApplier
	Autopilot            *agentpod.AutopilotControllerService // AutopilotController automation service
	Channel              *channel.Service
	Ticket               *ticket.Service
	MRSync               *ticket.MRSyncService       // MR sync for webhook events
	AgentPodSettings     *agentpod.SettingsService   // AgentPod user settings
	AgentPodAIProvider   *agentpod.AIProviderService // AgentPod AI provider management
	AIResource           *airesourcesvc.Service      // Unified AI resource center
	VirtualKey           *virtualkeysvc.Service      // Virtual API keys (quota/billing handles)
	TokenQuota           *tokenquotasvc.Service      // Token quotas + usage report
	EnvBundle            *envbundlesvc.Service       // Env bundles for harness credential injection
	Billing              *billing.Service
	Message              *MessageService                  // Agent-to-agent messaging
	Hub                  *websocket.Hub                   // WebSocket hub for real-time communication
	EventBus             *eventbus.EventBus               // Event bus for real-time events
	Email                email.Service                    // Email service
	Invitation           *invitation.Service              // Organization invitations
	PromoCode            *promocode.Service               // Promo code management
	APIKey               *apikeyservice.Service           // API key management for third-party access
	APIKeyAdapter        *apikeyservice.MiddlewareAdapter // API key middleware adapter
	File                 *fileservice.Service
	// NOTE: GitProvider and SSHKey services have been removed (moved to user-level settings)

	// gRPC/mTLS Runner registration handler (optional, only when PKI is enabled)
	GRPCRunnerHandler      *GRPCRunnerHandler
	RunnerGRPCAdapter      *grpcserver.GRPCRunnerAdapter
	AgentWorkbenchRepo     agentworkbenchdomain.PersistenceRepository
	AgentWorkbenchHub      *agentworkbenchsvc.DeltaHub
	AgentWorkbenchIngress  *agentworkbenchsvc.Ingress
	AgentWorkbenchCommands *agentworkbenchsvc.CommandDispatcher
	AgentSessions          *sessionsvc.Service
	EmbedTokens            *embedtoken.Service

	// Sandbox query service
	SandboxQueryService *runner.SandboxQueryService // Sandbox status query service
	SandboxFsService    *runner.SandboxFsService    // Sandbox filesystem ops

	// Upgrade command sender (gRPC adapter)
	UpgradeCommandSender runner.UpgradeCommandSender

	// Log upload services
	LogUploadSender  runner.LogUploadCommandSender
	LogUploadService *runnerlogservice.Service

	// Relay services for terminal data streaming
	RelayManager        *relay.Manager        // Relay server management
	RelayTokenGenerator *relay.TokenGenerator // Relay token generation
	RelayDNSService     *relay.DNSService     // Relay DNS management
	RelayACMEManager    *acme.Manager         // ACME certificate management for Relay TLS

	// GeoIP resolver for geo-aware relay selection
	GeoResolver geo.Resolver

	// Runner version checker (optional, checks GitHub Releases for latest version)
	VersionChecker *runner.VersionChecker

	// Extension services (Skills marketplace, MCP servers).
	// ExtensionRepo + MarketplaceWorker were dropped when admin skill
	// registries moved to Connect-RPC (commit-link). Connect handlers
	// read them straight from the serviceContainer, REST has no remaining
	// consumer, so keeping them here would be dead state.
	Extension *extensionservice.Service

	// Workflow services
	Workflow             *workflow.WorkflowService
	WorkflowRun          *workflow.WorkflowRunService
	WorkflowOrchestrator *workflow.WorkflowOrchestrator
	WorkflowScheduler    *workflow.WorkflowScheduler

	// Coordinator service (auto-harness integration: scans external task sources
	// → tickets → dispatches do-agent pods).
	Coordinator *coordinatorservice.Service

	Expert      *expertSvc.Service
	WorkerSpecs workerspecsvc.SnapshotRepository

	// Skill is the git-backed author-in-platform skill service (namespace
	// am-skills). Nil when git-backing / packager is not configured; routes
	// no-op. Additive to the external-import skill flow.
	Skill *skillSvc.Service

	IMBridge *imbridgesvc.Bridge

	PendingQueue *runner.PendingCommandQueue

	// SSO service for enterprise SSO integration
	SSO *ssoservice.Service

	// Support ticket service (user-level, no org scope)
	SupportTicket *supportticketservice.Service

	// Token usage service
	TokenUsage *tokenusagesvc.Service

	// Resource grant/sharing service
	Grant           *grantservice.Service
	PreviewSessions *previewservice.Service

	// Redis is optional — when non-nil, route-level rate limiters can use it.
	// Nil in tests or minimal deployments; middleware treats nil as no-op.
	Redis *redis.Client
}
