package omnigent

import (
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	authservice "github.com/anthropics/agentsmesh/backend/internal/service/auth"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	relayservice "github.com/anthropics/agentsmesh/backend/internal/service/relay"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
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
	Bridge          *EventBridge
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
	SessionFiles    *sessionfilesvc.Service
	Version         string
}
