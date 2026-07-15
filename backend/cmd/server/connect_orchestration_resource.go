package main

import (
	"net/http"

	"connectrpc.com/connect"

	orchestrationresourceconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/orchestration_resource"
)

func mountOrchestrationResourceService(
	mux *http.ServeMux,
	services *serviceContainer,
	options []connect.HandlerOption,
) {
	orchestrationresourceconnect.Mount(
		mux,
		orchestrationresourceconnect.NewServer(
			services.orchestration,
			services.org,
		),
		options...,
	)
}
