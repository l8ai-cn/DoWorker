package main

import (
	"net/http"

	"connectrpc.com/connect"
	agentworkbenchconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/agent_workbench"
	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	workbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
)

func mountAgentWorkbenchService(
	mux *http.ServeMux,
	services *serviceContainer,
	rest *v1.Services,
) {
	var repository agentworkbenchconnect.PersistenceRepository
	var executor agentworkbenchconnect.CommandExecutor
	if rest != nil && rest.AgentWorkbenchRepo != nil {
		repository = rest.AgentWorkbenchRepo
		executor = rest.AgentWorkbenchCommands
	}
	var sessions agentworkbenchconnect.SessionLookup
	var hub = restAgentWorkbenchHub(rest)
	if rest != nil {
		sessions = rest.AgentSessions
	}
	agentworkbenchconnect.Mount(
		mux,
		agentworkbenchconnect.NewServer(
			repository,
			hub,
			sessions,
			services.org,
			executor,
		),
		connect.WithInterceptors(agentworkbenchconnect.NewAuthInterceptor(
			services.auth.AccessTokenManager(),
			services.auth.AccessTokenAudience(),
			restAgentWorkbenchEmbedTokens(rest),
		)),
	)
}

func restAgentWorkbenchHub(rest *v1.Services) *workbenchsvc.DeltaHub {
	if rest == nil {
		return nil
	}
	return rest.AgentWorkbenchHub
}

func restAgentWorkbenchEmbedTokens(
	rest *v1.Services,
) agentworkbenchconnect.EmbedTokenValidator {
	if rest == nil {
		return nil
	}
	return rest.EmbedTokens
}
