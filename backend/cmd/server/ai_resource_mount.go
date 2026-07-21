package main

import (
	"net/http"

	"connectrpc.com/connect"

	airesourceconnect "github.com/l8ai-cn/agentcloud/backend/internal/api/connect/ai_resource"
)

func mountAIResourceService(mux *http.ServeMux, services *serviceContainer, options []connect.HandlerOption) {
	airesourceconnect.Mount(mux, airesourceconnect.NewServer(services.aiResource, services.org), options...)
}
