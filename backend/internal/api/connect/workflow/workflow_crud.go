package workflowconnect

import (
	"context"
	"encoding/json"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

func jsonRawFromString(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}

// CreateWorkflow — REST analogue: POST /workflows.
func (s *Server) CreateWorkflow(
	ctx context.Context, req *connect.Request[workflowv1.CreateWorkflowRequest],
) (*connect.Response[workflowv1.Workflow], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	return nil, connect.NewError(
		connect.CodeFailedPrecondition,
		errors.New("workflow definitions must be created through orchestration validate-plan-apply"),
	)
}
