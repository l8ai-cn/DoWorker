package main

import (
	"net/http"

	"connectrpc.com/connect"

	goalloopconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/goalloop"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
)

func mountGoalLoopService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.goalLoop == nil {
		return
	}
	drafts := goalloopsvc.NewDraftGenerator(
		svc.aiResource,
		svc.workerDraftGenerator,
	)
	server := goalloopconnect.NewServer(
		svc.goalLoop,
		svc.org,
		goalloopconnect.WithAIGeneration(drafts),
	)
	goalloopconnect.Mount(mux, server, opts...)
}
