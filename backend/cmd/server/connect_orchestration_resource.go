package main

import (
	"net/http"

	"connectrpc.com/connect"

	orchestrationresourceconnect "github.com/l8ai-cn/agentcloud/backend/internal/api/connect/orchestration_resource"
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
			services.bindingApply,
			services.workerTemplateApply,
			services.workerApply,
			services.promptApply,
			services.expertApply,
			services.workflowApply,
			services.goalLoopApply,
			services.org,
		),
		options...,
	)
}
