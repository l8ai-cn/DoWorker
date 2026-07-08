package sessionapi

import (
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
	envbundlesvc "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	authservice "github.com/anthropics/agentsmesh/backend/internal/service/auth"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	relayservice "github.com/anthropics/agentsmesh/backend/internal/service/relay"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	commentsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessioncomment"
	permgrantsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionpermission"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	userservice "github.com/anthropics/agentsmesh/backend/internal/service/user"
)

type Deps struct {
	JWTSecret       string
	Auth            *authservice.Service
	User            *userservice.Service
	Org             *organization.Service
	Agent           *agentservice.AgentService
	Runner          *runnerservice.Service
	Sessions        *sessionsvc.Service
	Items           *itemsvc.Service
	Hub             *SessionHub
	Updates         *SessionUpdatesHub
	Elicitations    *ElicitationStore
	Stream          *SessionStreamPublisher
	PodOrchestrator *agentpod.PodOrchestrator
	Pod             *agentpod.PodService
	PodCoordinator  *runnerservice.PodCoordinator
	CommandSender   runnerservice.RunnerCommandSender
	RelayManager    *relayservice.Manager
	RelayTokens     *relayservice.TokenGenerator
	SessionUsage    *sessionusagesvc.Service
	Policies        *permissionpolicysvc.Service
	ReadState       *ReadStateStore
	SandboxFs       *runnerservice.SandboxFsService
	SessionFiles      *sessionfilesvc.Service
	SessionComments   *commentsvc.Service
	SessionPermissions *permgrantsvc.Service
	Grants             *grantservice.Service
	AIModels           *aimodelsvc.Service
	EnvBundles         *envbundlesvc.Service
	VirtualKeys        *virtualkeysvc.Service
	TokenQuotas        *tokenquotasvc.Service
	Version           string
}
