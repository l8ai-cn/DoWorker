package sessionapi

import (
	"context"

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
)

type sessionPromptOutbox interface {
	PersistAndQueue(context.Context, sessionmessagesvc.PromptInput) error
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
	PodOrchestrator    *agentpod.PodOrchestrator
	Pod                *agentpod.PodService
	PodCoordinator     *runnerservice.PodCoordinator
	CommandSender      runnerservice.RunnerCommandSender
	RelayManager       *relayservice.Manager
	RelayTokens        *relayservice.TokenGenerator
	SessionUsage       *sessionusagesvc.Service
	Policies           *permissionpolicysvc.Service
	ReadState          *ReadStateStore
	SandboxFs          *runnerservice.SandboxFsService
	SessionFiles       *sessionfilesvc.Service
	MessageOutbox      sessionPromptOutbox
	SessionComments    *commentsvc.Service
	SessionPermissions *permgrantsvc.Service
	Grants             *grantservice.Service
	AIResources        ModelResourceLister
	EnvBundles         *envbundlesvc.Service
	VirtualKeys        *virtualkeysvc.Service
	TokenQuotas        *tokenquotasvc.Service
	Version            string
}
