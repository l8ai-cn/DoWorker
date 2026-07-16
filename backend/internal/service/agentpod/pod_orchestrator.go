package agentpod

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/agentfile"
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	userService "github.com/anthropics/agentsmesh/backend/internal/service/user"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

var (
	ErrMissingRunnerID                  = errors.New("runner_id is required")
	ErrMissingAgentSlug                 = errors.New("agent_slug is required")
	ErrMissingAgentAdapter              = errors.New("agent adapter_id is required")
	ErrSourcePodNotFound                = errors.New("source pod not found")
	ErrSourcePodAccessDenied            = errors.New("source pod belongs to different organization")
	ErrSourcePodNotTerminated           = errors.New("source pod is not terminated")
	ErrSourcePodAlreadyResumed          = errors.New("source pod already resumed")
	ErrResumeAgentMismatch              = errors.New("resume requires same agent as source pod")
	ErrResumeRunnerMismatch             = errors.New("resume requires same runner")
	ErrResumeLineageInvalid             = errors.New("resume source lineage is invalid")
	ErrConfigBuildFailed                = errors.New("failed to build pod configuration")
	ErrInvalidAgentfileLayer            = errors.New("invalid agentfile layer")
	ErrRunnerDispatchFailed             = errors.New("failed to dispatch pod to runner")
	ErrUnsupportedInteractionMode       = errors.New("agent type does not support the requested interaction mode")
	ErrMissingModelResource             = errors.New("model_resource_id is required for this agent")
	ErrModelResourceResolverUnavailable = errors.New("model resource resolver is not configured")
	ErrModelResourceEnvConflict         = errors.New("model resource conflicts with an existing runtime environment value")
	ErrModelResourceCommandConflict     = errors.New("model resource conflicts with an existing model launch argument")
	ErrWorkerSpecModelChanged           = errors.New("model resource changed after workerspec resolution")
)

const (
	errCodeRunnerUnreachable = "RUNNER_UNREACHABLE"
	errCodeQueueFull         = "QUEUE_FULL"
)

type PodCoordinatorForOrchestrator interface {
	CreatePod(ctx context.Context, runnerID int64, cmd *runnerv1.CreatePodCommand) error
	CreatePodOrQueue(ctx context.Context, runnerID int64, cmd *runnerv1.CreatePodCommand, opts podDomain.CreatePodQueueOpts) error
}

type BillingServiceForOrchestrator interface {
	CheckQuota(ctx context.Context, orgID int64, quotaType string, amount int) error
}

type UserServiceForOrchestrator interface {
	GetDefaultGitCredential(ctx context.Context, userID int64) (*user.GitCredential, error)
	GetDecryptedCredentialToken(ctx context.Context, userID, credentialID int64) (*userService.DecryptedCredential, error)
}

type RepositoryServiceForOrchestrator interface {
	GetAccessibleByID(ctx context.Context, id, orgID, userID int64) (*gitprovider.Repository, error)
	FindAccessibleByOrgSlug(ctx context.Context, orgID, userID int64, slug string) (*gitprovider.Repository, error)
}

type TicketServiceForOrchestrator interface {
	GetTicket(ctx context.Context, ticketID int64) (*ticket.Ticket, error)
	GetTicketBySlug(ctx context.Context, organizationID int64, slug string) (*ticket.Ticket, error)
}

type RunnerSelectorForOrchestrator interface {
	SelectRunnerWithAffinity(ctx context.Context, orgID int64, userID int64, agentSlug string, hints *runnerDomain.AffinityHints, repoHistory map[int64]int) (*runnerDomain.Runner, error)
	ResolveRunnerForCreate(ctx context.Context, runnerID, orgID, userID int64, agentSlug string, allowUnavailable bool) (*runnerDomain.Runner, error)
}

type RunnerQueryForOrchestrator interface {
	GetRunner(ctx context.Context, runnerID int64) (*runnerDomain.Runner, error)
}

type AgentResolverForOrchestrator interface {
	GetAgent(ctx context.Context, slug string) (*agentDomain.Agent, error)
}

type UserConfigQueryForOrchestrator interface {
	GetUserConfigPrefs(ctx context.Context, userID int64, agentSlug string) map[string]interface{}
}

// KnowledgeBaseResolverForOrchestrator resolves KB mounts for pod creation.
// Nil means the KB feature is disabled (internal Gitea not configured).
type KnowledgeBaseResolverForOrchestrator interface {
	ResolveMountsForPod(ctx context.Context, orgID int64, agentSlug string, requested []kbservice.MountRequest) ([]*kbservice.ResolvedMount, error)
	CloneToken() string
}

type PodOrchestratorDeps struct {
	PodService       *PodService
	ConfigBuilder    *agent.ConfigBuilder
	PodCoordinator   PodCoordinatorForOrchestrator
	BillingService   BillingServiceForOrchestrator
	UserService      UserServiceForOrchestrator
	RepoService      RepositoryServiceForOrchestrator
	TicketService    TicketServiceForOrchestrator
	RunnerSelector   RunnerSelectorForOrchestrator
	AgentResolver    AgentResolverForOrchestrator
	RunnerQuery      RunnerQueryForOrchestrator
	UserConfigQuery  UserConfigQueryForOrchestrator
	PodRepo          podDomain.PodRepository
	PermissionPolicy *permissionpolicysvc.Service
	KnowledgeBases   KnowledgeBaseResolverForOrchestrator
	ModelResources   ModelResourceResolver
	WorkerCreation   WorkerCreationPreparer
	WorkerSpecs      WorkerSpecSnapshotLoader
}

type PodOrchestrator struct {
	podService       *PodService
	configBuilder    *agent.ConfigBuilder
	podCoordinator   PodCoordinatorForOrchestrator
	billingService   BillingServiceForOrchestrator
	userService      UserServiceForOrchestrator
	repoService      RepositoryServiceForOrchestrator
	ticketService    TicketServiceForOrchestrator
	runnerSelector   RunnerSelectorForOrchestrator
	agentResolver    AgentResolverForOrchestrator
	runnerQuery      RunnerQueryForOrchestrator
	userConfigQuery  UserConfigQueryForOrchestrator
	podRepo          podDomain.PodRepository
	permissionPolicy *permissionpolicysvc.Service
	knowledgeBases   KnowledgeBaseResolverForOrchestrator
	modelResources   ModelResourceResolver
	workerCreation   WorkerCreationPreparer
	workerSnapshots  WorkerSnapshotPreparer
	workerSpecs      WorkerSpecSnapshotLoader
}

type agentfileResolved struct {
	InteractionMode       string
	BranchName            string
	RepositoryID          *int64
	Repository            *gitprovider.Repository
	Prompt                string
	Knowledge             []agentfile.KnowledgeSpec
	MergedAgentfileSource string
	ConfigValues          agentDomain.ConfigValues
}

func NewPodOrchestrator(deps *PodOrchestratorDeps) *PodOrchestrator {
	workerSnapshots, _ := deps.WorkerCreation.(WorkerSnapshotPreparer)
	return &PodOrchestrator{
		podService:       deps.PodService,
		configBuilder:    deps.ConfigBuilder,
		podCoordinator:   deps.PodCoordinator,
		billingService:   deps.BillingService,
		userService:      deps.UserService,
		repoService:      deps.RepoService,
		ticketService:    deps.TicketService,
		runnerSelector:   deps.RunnerSelector,
		agentResolver:    deps.AgentResolver,
		runnerQuery:      deps.RunnerQuery,
		userConfigQuery:  deps.UserConfigQuery,
		podRepo:          deps.PodRepo,
		permissionPolicy: deps.PermissionPolicy,
		knowledgeBases:   deps.KnowledgeBases,
		modelResources:   deps.ModelResources,
		workerCreation:   deps.WorkerCreation,
		workerSnapshots:  workerSnapshots,
		workerSpecs:      deps.WorkerSpecs,
	}
}
