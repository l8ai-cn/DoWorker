package sessionapi

import (
	"context"

	agentworkbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	authservice "github.com/anthropics/agentsmesh/backend/internal/service/auth"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	envbundlesvc "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	relayservice "github.com/anthropics/agentsmesh/backend/internal/service/relay"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	commentsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessioncomment"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	permgrantsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionpermission"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	userservice "github.com/anthropics/agentsmesh/backend/internal/service/user"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
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
	Pod                *agentpod.PodService
	PodCoordinator     *runnerservice.PodCoordinator
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
