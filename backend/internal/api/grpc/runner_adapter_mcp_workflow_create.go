package grpc

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

func (a *GRPCRunnerAdapter) mcpCreateWorkflow(
	context.Context,
	*middleware.TenantContext,
	string,
	[]byte,
) (interface{}, *mcpError) {
	return nil, newMcpError(
		409,
		"workflow definitions must be created through orchestration validate-plan-apply",
	)
}
