package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type mcpWorkerPlanPayload struct {
	PlanID string `json:"plan_id"`
}

func (a *GRPCRunnerAdapter) mcpCreatePod(
	ctx context.Context,
	tenant *middleware.TenantContext,
	payload []byte,
) (interface{}, *mcpError) {
	params, decodeErr := decodeMCPWorkerPlanPayload(payload)
	if decodeErr != nil {
		return nil, decodeErr
	}
	if a.workerPlanAuthorizer == nil ||
		a.workerPlanApplier == nil ||
		a.workerPodReader == nil {
		return nil, newMcpError(503, "Worker apply service is unavailable")
	}
	scope := control.Scope{
		OrganizationID: tenant.OrganizationID,
		OrganizationSlug: slugkit.Slug(
			tenant.OrganizationSlug,
		),
		ActorID: tenant.UserID,
	}
	if err := scope.Validate(); err != nil {
		return nil, newMcpError(500, "organization scope is unavailable")
	}
	if err := a.workerPlanAuthorizer.AuthorizeApply(
		ctx,
		scope,
		params.PlanID,
	); err != nil {
		return nil, mapWorkerPlanErrorToMCP(err)
	}
	applied, err := a.workerPlanApplier.Apply(ctx, scope, params.PlanID)
	if err != nil {
		return nil, mapWorkerPlanErrorToMCP(err)
	}
	pod, err := a.workerPodReader.GetPod(ctx, applied.PodKey)
	if err != nil || pod == nil || pod.PodKey != applied.PodKey {
		return nil, newMcpError(500, "failed to load applied Worker Pod")
	}
	return map[string]interface{}{
		"pod": map[string]interface{}{
			"pod_key": pod.PodKey,
			"status":  pod.Status,
		},
	}, nil
}

func decodeMCPWorkerPlanPayload(
	payload []byte,
) (mcpWorkerPlanPayload, *mcpError) {
	var params mcpWorkerPlanPayload
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&params); err != nil {
		return params, newMcpError(400, "invalid request payload")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return params, newMcpError(400, "invalid request payload")
	}
	parsed, err := uuid.Parse(params.PlanID)
	if err != nil || parsed == uuid.Nil || parsed.String() != params.PlanID {
		return params, newMcpError(400, "plan_id must be a canonical UUID")
	}
	return params, nil
}

func mapWorkerPlanErrorToMCP(err error) *mcpError {
	switch {
	case errors.Is(err, control.ErrInvalid):
		return newMcpError(400, "Worker plan is invalid")
	case errors.Is(err, controlservice.ErrForbidden):
		return newMcpError(403, "Worker plan access is forbidden")
	case errors.Is(err, control.ErrNotFound):
		return newMcpError(404, "Worker plan was not found")
	case errors.Is(err, control.ErrConflict),
		errors.Is(err, control.ErrStale),
		errors.Is(err, control.ErrExpired),
		errors.Is(err, control.ErrConsumed):
		return newMcpError(409, "Worker plan state changed; create a new plan")
	case errors.Is(err, podDomain.ErrQueueFull):
		return newMcpError(429, "Runner pending queue is full")
	case errors.Is(err, agentsvc.ErrNoAvailableRunner):
		return newMcpError(422, "No runner is available for the Worker snapshot")
	case errors.Is(err, controlservice.ErrUnavailable):
		return newMcpError(503, "Worker apply service is unavailable")
	default:
		return newMcpError(500, "failed to apply Worker plan")
	}
}
