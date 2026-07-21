package sessionapi

import (
	"context"
	"time"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessionDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	agentworkbenchdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	agentservice "github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	sessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	authservice "github.com/l8ai-cn/agentcloud/backend/internal/service/auth"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	envbundlesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
	grantservice "github.com/l8ai-cn/agentcloud/backend/internal/service/grant"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	permissionpolicysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/permissionpolicy"
	relayservice "github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	commentsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessioncomment"
	sessionfilesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionfile"
	sessionmessagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionmessage"
	permgrantsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionpermission"
	sessionusagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionusage"
	tokenquotasvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenquota"
	userservice "github.com/l8ai-cn/agentcloud/backend/internal/service/user"
	virtualkeysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/virtualkey"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/embedtoken"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type sessionPromptOutbox interface {
	PersistAndQueue(context.Context, sessionmessagesvc.PromptInput) error
}

type previewSessionRevoker interface {
	RevokeUser(context.Context, int64) error
}

type sandboxFilesystem interface {
	IsConnected(runnerID int64) bool
	Exec(
		context.Context,
		int64,
		*runnerv1.SandboxFsCommand,
	) (*runnerv1.SandboxFsResultEvent, error)
}

type sessionPodOrchestrator interface {
	CreatePod(
		context.Context,
		*agentpod.OrchestrateCreatePodRequest,
	) (*agentpod.OrchestrateCreatePodResult, error)
	DispatchDeferredPod(
		context.Context,
		*agentpod.OrchestrateCreatePodRequest,
		*agentpod.OrchestrateCreatePodResult,
	) (*agentpod.OrchestrateCreatePodResult, error)
}

type sessionWorkerDraftFactory interface {
	NewFreshPodDraft(
		context.Context,
		specservice.Scope,
		workercreation.FreshPodDraftInput,
	) (workercreation.Draft, error)
}

type sessionPodLifecycle interface {
	TerminatePod(context.Context, string) error
	TerminatePodDeleteBranch(context.Context, string) error
}

type sessionDeferredCommitter interface {
	CommitCreate(
		context.Context,
		*sessionDomain.Session,
		*podDomain.PendingCommand,
		int,
		func(*itemsvc.Service) error,
	) error
}

type sessionDispatchQueue interface {
	AllowsDurableCommand(int64) bool
	MaxPerRunner() int
	SendPromptTTL() time.Duration
	SealPayload([]byte) ([]byte, error)
	TriggerDrain(int64)
}

type Deps struct {
	Auth               *authservice.Service
	User               *userservice.Service
	Org                *organization.Service
	Agent              *agentservice.AgentService
	Runner             *runnerservice.Service
	Sessions           *sessionsvc.Service
	Items              *itemsvc.Service
	Hub                *SessionHub
	Updates            *SessionUpdatesHub
	Elicitations       *ElicitationStore
	Stream             *SessionStreamPublisher
	PodOrchestrator    sessionPodOrchestrator
	WorkerCreation     sessionWorkerDraftFactory
	Pod                *agentpod.PodService
	DeferredCommitter  sessionDeferredCommitter
	DispatchQueue      sessionDispatchQueue
	PodCoordinator     sessionPodLifecycle
	CommandSender      runnerservice.RunnerCommandSender
	RelayManager       *relayservice.Manager
	RelayTokens        *relayservice.TokenGenerator
	SessionUsage       *sessionusagesvc.Service
	Policies           *permissionpolicysvc.Service
	ReadState          *ReadStateStore
	SandboxFs          sandboxFilesystem
	SessionFiles       *sessionfilesvc.Service
	WorkbenchRepo      agentworkbenchdomain.Repository
	MessageOutbox      sessionPromptOutbox
	SessionComments    *commentsvc.Service
	SessionPermissions *permgrantsvc.Service
	Grants             *grantservice.Service
	AIResources        ModelResourceLister
	EnvBundles         *envbundlesvc.Service
	VirtualKeys        *virtualkeysvc.Service
	TokenQuotas        *tokenquotasvc.Service
	EmbedTokens        *embedtoken.Service
	PreviewSessions    previewSessionRevoker
	Version            string
}
